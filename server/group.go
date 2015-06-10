package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Pattern struct {
	Blocks      []BlockLedger      `json:"blocks"`
	Connections []ConnectionLedger `json:"connections"`
	Groups      []Group            `json:"groups"`
	Sources     []SourceLedger     `json:"sources"`
	Links       []LinkLedger       `json:"links"`
}

type Node interface {
	GetID() int
	GetParent() *Group
	SetParent(*Group)
}

type Group struct {
	Id       int      `json:"id"`
	Label    string   `json:"label"`
	Children []int    `json:"children"`
	Parent   *Group   `json:"-"`
	Position Position `json:"position"`
}

type ProtoGroup struct {
	Group    int      `json:"parent"`
	Children []int    `json:"children"`
	Label    string   `json:"label"`
	Position Position `json:"position"`
}

func (g *Group) GetID() int {
	return g.Id
}

func (g *Group) GetParent() *Group {
	return g.Parent
}

func (g *Group) SetParent(group *Group) {
	g.Parent = group
}

func (s *Server) ListGroups() []Group {
	groups := []Group{}
	for _, g := range s.groups {
		groups = append(groups, *g)
	}
	return groups
}

func (s *Server) GroupIndexHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.ListGroups(), http.StatusOK)
}

func (s *Server) DetachChild(g Node) error {
	parent := g.GetParent()
	if parent == nil {
		return errors.New("no parent to detach from")
	}

	id := g.GetID()

	child := -1
	for i, v := range parent.Children {
		if v == id {
			child = i
		}
	}

	if child == -1 {
		return errors.New("could not remove child from group: child does not exist")
	}

	parent.Children = append(parent.Children[:child], parent.Children[child+1:]...)

	s.websocketBroadcast(Update{Action: DELETE, Type: CHILD, Data: wsGroupChild{
		Group: wsId{parent.GetID()},
		Child: wsId{g.GetID()},
	}})

	return nil
}

func (s *Server) AddChildToGroup(id int, n Node) error {
	newParent, ok := s.groups[id]
	if !ok {
		return errors.New("group not found")
	}

	nid := n.GetID()
	for _, v := range newParent.Children {
		if v == nid {
			return errors.New("node already child of this group")
		}
	}

	newParent.Children = append(newParent.Children, nid)
	if n.GetParent() != nil {
		err := s.DetachChild(n)
		if err != nil {
			return err
		}
	}

	n.SetParent(newParent)

	s.websocketBroadcast(Update{Action: CREATE, Type: CHILD, Data: wsGroupChild{
		Group: wsId{id},
		Child: wsId{nid},
	}})
	return nil
}

// CreateGroupHandler responds to a POST request to instantiate a new group and add it to the Server.
// Moves all of the specified children out of the parent's group and into the new group.
func (s *Server) GroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, Error{"could not read request body"}, http.StatusBadRequest)
		return
	}

	var g ProtoGroup
	err = json.Unmarshal(body, &g)
	if err != nil {
		writeJSON(w, Error{"could not read JSON"}, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()

	newGroup, err := s.CreateGroup(g)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, newGroup, http.StatusOK)
}

func (s *Server) CreateGroup(g ProtoGroup) (*Group, error) {
	newGroup := &Group{
		Children: []int{},
		Label:    g.Label,
		Position: g.Position,
		Id:       s.GetNextID(),
	}

	//if newGroup.Children == nil {
	//	newGroup.Children = []int{}
	//}

	for _, c := range g.Children {
		_, okb := s.blocks[c]
		_, okg := s.groups[c]
		_, oks := s.sources[c]
		if !okb && !okg && !oks {
			return nil, errors.New("could not create group: invalid children")
		}
	}

	s.groups[newGroup.Id] = newGroup
	s.websocketBroadcast(Update{Action: CREATE, Type: GROUP, Data: wsGroup{*newGroup}})

	err := s.AddChildToGroup(g.Group, newGroup)
	if err != nil {
		return nil, err
	}

	for _, c := range g.Children {
		if cb, ok := s.blocks[c]; ok {
			err = s.AddChildToGroup(newGroup.Id, cb)
		}
		if cg, ok := s.groups[c]; ok {
			err = s.AddChildToGroup(newGroup.Id, cg)
		}
		if cs, ok := s.sources[c]; ok {
			err = s.AddChildToGroup(newGroup.Id, cs)
		}
		if err != nil {
			return nil, err
		}
	}

	return newGroup, nil
}

func (s *Server) DeleteGroup(id int) error {
	group, ok := s.groups[id]
	if !ok {
		return errors.New("could not find group to delete")
	}

	for _, c := range group.Children {
		if _, ok := s.blocks[c]; ok {
			err := s.DeleteBlock(c)
			if err != nil {
				return err
			}
		} else if _, ok := s.groups[c]; ok {
			err := s.DeleteGroup(c)
			if err != nil {
				return err
			}
		}
	}

	s.DetachChild(group)
	delete(s.groups, id)
	s.websocketBroadcast(Update{Action: DELETE, Type: GROUP, Data: wsGroup{wsId{id}}})
	return nil
}

func (s *Server) GroupDeleteHandler(w http.ResponseWriter, r *http.Request) {

	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()

	err = s.DeleteGroup(id)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// returns a description of the group - its id and childreen
func (s *Server) GroupHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}
	s.Lock()
	defer s.Unlock()
	group, ok := s.groups[id]
	if !ok {
		writeJSON(w, Error{"could not find group"}, http.StatusBadRequest)
		return
	}
	writeJSON(w, group, http.StatusOK)
}

func (s *Server) GroupExportHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()
	p, err := s.Export(id)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, p, http.StatusOK)
}

func (s *Server) ExportGroup(id int) (*Pattern, error) {
	p := &Pattern{
		Blocks:      []BlockLedger{},
		Sources:     []SourceLedger{},
		Groups:      []Group{},
		Connections: []ConnectionLedger{},
		Links:       []LinkLedger{},
	}

	g, ok := s.groups[id]
	if !ok {
		return nil, errors.New("could not find group to export")
	}

	p.Groups = append(p.Groups, *g)
	for _, c := range g.Children {
		if b, ok := s.blocks[c]; ok {
			p.Blocks = append(p.Blocks, *b)
			continue
		}
		if source, ok := s.sources[c]; ok {
			p.Sources = append(p.Sources, *source)
			continue
		}
		if group, ok := s.groups[c]; ok {
			g, err := s.ExportGroup(group.Id)
			if err != nil {
				return nil, err
			}

			p.Blocks = append(p.Blocks, g.Blocks...)
			p.Groups = append(p.Groups, g.Groups...)
			p.Sources = append(p.Sources, g.Sources...)
			continue
		}
	}
	return p, nil
}

func (s *Server) Export(id int) (*Pattern, error) {
	p, err := s.ExportGroup(id)
	if err != nil {
		return nil, err
	}

	ids := make(map[int]struct{})
	for _, b := range p.Blocks {
		ids[b.Id] = struct{}{}
	}

	for _, g := range p.Groups {
		ids[g.Id] = struct{}{}
	}

	for _, source := range p.Sources {
		ids[source.Id] = struct{}{}
	}

	for _, c := range s.connections {
		_, sourceIncluded := ids[c.Source.Id]
		_, targetIncluded := ids[c.Target.Id]
		if sourceIncluded && targetIncluded {
			p.Connections = append(p.Connections, *c)
		}
	}

	for _, l := range s.links {
		_, sourceIncluded := ids[l.Block.Id]
		_, targetIncluded := ids[l.Source.Id]
		if sourceIncluded && targetIncluded {
			p.Links = append(p.Links, *l)
		}
	}

	return p, nil
}

func (s *Server) GroupImportHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}

	var p Pattern
	err = json.Unmarshal(body, &p)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()

	err = s.ImportGroup(id, p)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ImportGroup(id int, p Pattern) error {
	parents := make(map[int]int) // old child id / old parent id
	newIds := make(map[int]int)  // old id / new id

	if _, ok := s.groups[id]; !ok {
		return errors.New("could not attach to group: does not exist")
	}

	for _, g := range p.Groups {
		ng, err := s.CreateGroup(ProtoGroup{
			Label:    g.Label,
			Position: g.Position,
		})

		if err != nil {
			return err
		}

		newIds[g.Id] = ng.Id

		for _, c := range g.Children {
			parents[c] = g.Id
		}
	}

	for _, b := range p.Blocks {
		nb, err := s.CreateBlock(ProtoBlock{
			Label:    b.Label,
			Position: b.Position,
			Type:     b.Type,
		})

		if err != nil {
			return err
		}

		newIds[b.Id] = nb.Id
	}

	for _, source := range p.Sources {
		ns, err := s.CreateSource(ProtoSource{
			Label:    source.Label,
			Position: source.Position,
			Type:     source.Type,
		})

		if err != nil {
			return err
		}

		newIds[source.Id] = ns.Id
	}

	for _, c := range p.Connections {
		c.Source.Id = newIds[c.Source.Id]
		c.Target.Id = newIds[c.Target.Id]
		_, err := s.CreateConnection(ProtoConnection{
			Source: c.Source,
			Target: c.Target,
		})
		if err != nil {
			return err
		}
	}

	for _, l := range p.Links {
		pl := ProtoLink{}
		pl.Block.Id = newIds[l.Block.Id]
		pl.Source.Id = newIds[l.Source.Id]
		_, err := s.CreateLink(pl)
		if err != nil {
			return err
		}
	}

	for _, source := range p.Sources {
		if source.Parameters != nil {
			err := s.ModifySource(newIds[source.Id], source.Parameters)
			if err != nil {
				return err
			}
		}
	}

	for _, b := range p.Blocks {
		for route, v := range b.Inputs {
			err := s.ModifyBlockRoute(newIds[b.Id], route, v.Value)
			if err != nil {
				return err
			}
		}
	}

	for _, g := range p.Groups {
		for _, c := range g.Children {
			var n Node
			if bn, ok := s.blocks[newIds[c]]; ok {
				n = bn
			}
			if bg, ok := s.groups[newIds[c]]; ok {
				n = bg
			}
			if bs, ok := s.sources[newIds[c]]; ok {
				n = bs
			}
			if n == nil {
				return errors.New("could not add node, node does not exist")
			}

			err := s.AddChildToGroup(newIds[g.Id], n)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) GroupModifyLabelHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, Error{"could not read request body"}, http.StatusBadRequest)
		return
	}

	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}

	var l string
	err = json.Unmarshal(body, &l)
	if err != nil {
		writeJSON(w, Error{"could not unmarshal: " + string(body) + ""}, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()

	g, ok := s.groups[id]
	if !ok {
		writeJSON(w, Error{"no block found"}, http.StatusBadRequest)
		return
	}

	g.Label = l

	s.websocketBroadcast(Update{Action: UPDATE, Type: GROUP, Data: wsGroup{wsLabel{wsId{id}, l}}})

	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) GroupModifyAllChildrenHandler(w http.ResponseWriter, r *http.Request) {
}
func (s *Server) GroupModifyChildHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := getIDFromMux(vars)
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}

	childs, ok := vars["node_id"]
	if !ok {
		writeJSON(w, Error{"no ID supplied"}, http.StatusBadRequest)
		return
	}

	child, err := strconv.Atoi(childs)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	if id == child {
		writeJSON(w, Error{"cannot add group as member of itself"}, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()

	var n Node

	if _, ok := s.groups[id]; !ok {
		writeJSON(w, Error{"could not find id"}, http.StatusBadRequest)
		return
	}

	if b, ok := s.blocks[child]; ok {
		n = b
	}

	if g, ok := s.groups[child]; ok {
		n = g
	}

	if so, ok := s.sources[child]; ok {
		n = so
	}

	if n == nil {
		writeJSON(w, Error{"could not find id"}, http.StatusBadRequest)
		return
	}

	err = s.AddChildToGroup(id, n)
	if err != nil {
		writeJSON(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GroupPositionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromMux(mux.Vars(r))
	if err != nil {
		writeJSON(w, err, http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, Error{"could not read request body"}, http.StatusBadRequest)
		return
	}

	var p Position
	err = json.Unmarshal(body, &p)
	if err != nil {
		writeJSON(w, Error{"could not read JSON"}, http.StatusBadRequest)
		return
	}

	s.Lock()
	defer s.Unlock()

	g, ok := s.groups[id]
	if !ok {
		writeJSON(w, Error{"could not find group"}, http.StatusBadRequest)
		return
	}

	g.Position = p

	s.websocketBroadcast(Update{Action: UPDATE, Type: GROUP, Data: wsGroup{wsPosition{wsId{id}, p}}})
	w.WriteHeader(http.StatusNoContent)
}
