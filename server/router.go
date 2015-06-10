package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) NewRouter() http.Handler {

	var routes = []struct {
		Name        string
		Pattern     string
		Method      string
		HandlerFunc http.HandlerFunc
	}{
		{"UpdateSocket", "/updates", "GET", s.UpdateSocketHandler},
		{"BlockLibrary", "/blocks/library", "GET", s.BlockLibraryHandler},
		{"SourceLibrary", "/sources/library", "GET", s.SourceLibraryHandler},
		{"GroupIndex", "/groups", "GET", s.GroupIndexHandler},
		{"Group", "/groups/{id}", "GET", s.GroupHandler},
		{"GroupCreate", "/groups", "POST", s.GroupCreateHandler},
		{"GroupExport", "/groups/{id}/export", "GET", s.GroupExportHandler},
		{"GroupImport", "/groups/{id}/import", "POST", s.GroupImportHandler},
		{"GroupModifyLabel", "/groups/{id}/label", "PUT", s.GroupModifyLabelHandler},
		{"GroupModifyAllChildren", "/groups/{id}/children", "PUT", s.GroupModifyAllChildrenHandler},
		{"GroupModifyChild", "/groups/{id}/children/{node_id}", "PUT", s.GroupModifyChildHandler},
		{"GroupPosition", "/groups/{id}/position", "PUT", s.GroupPositionHandler},
		{"GroupDelete", "/groups/{id}", "DELETE", s.GroupDeleteHandler},
		{"BlockIndex", "/blocks", "GET", s.BlockIndexHandler},
		{"Block", "/blocks/{id}", "GET", s.BlockHandler},
		{"BlockCreate", "/blocks", "POST", s.BlockCreateHandler},
		{"BlockDelete", "/blocks/{id}", "DELETE", s.BlockDeleteHandler},
		{"BlockModifyName", "/blocks/{id}/label", "PUT", s.BlockModifyNameHandler},
		{"BlockModifyRoute", "/blocks/{id}/routes/{index}", "PUT", s.BlockModifyRouteHandler},
		{"BlockModifyPosition", "/blocks/{id}/position", "PUT", s.BlockModifyPositionHandler},
		{"ConnectionIndex", "/connections", "GET", s.ConnectionIndexHandler},
		{"Connection", "/connections/{id}", "GET", s.ConnectionHandler},
		{"ConnectionCreate", "/connections", "POST", s.ConnectionCreateHandler},
		{"ConnectionModifyCoordinates", "/connections/{id}/coordinates", "PUT", s.ConnectionModifyCoordinates},
		{"ConnectionDelete", "/connections/{id}", "DELETE", s.ConnectionDeleteHandler},
		{"SourceCreate", "/sources", "POST", s.SourceCreateHandler},
		{"SourceIndex", "/sources", "GET", s.SourceIndexHandler},
		{"SourceModifyName", "/sources/{id}/label", "PUT", s.SourceModifyNameHandler},
		{"SourceModifyPosition", "/sources/{id}/position", "PUT", s.SourceModifyPositionHandler},
		{"SourceModify", "/sources/{id}/params", "PUT", s.SourceModifyHandler},
		{"SourceGetValue", "/sources/{id}/value", "GET", s.SourceGetValueHandler},
		{"SourceSetValue", "/sources/{id}/value", "PUT", s.SourceSetValueHandler},
		{"Source", "/sources/{id}", "GET", s.SourceHandler},
		{"Source", "/sources/{id}", "DELETE", s.SourceDeleteHandler},
		{"LinkIndex", "/links", "GET", s.LinkIndexHandler},
		{"LinkCreate", "/links", "POST", s.LinkCreateHandler},
		{"LinkDelete", "/links/{id}", "DELETE", s.LinkDeleteHandler},
	}

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	return router

}
