var app = app || {};

// TODO:
// create a standard model API that the rest of the components can use
// this standard API should use WS to communicate back to server

(function() {
    'use strict';

    var dm = new app.Utils.DebounceManager();

    app.CoreModel = function() {
        this.entities = {};
        this.list = [];
        this.groups = [];
        this.edges = [];
        this.onChanges = [];

        this.focusedGroup = 0; // the current group in focus
        this.focusedNodes = []; // nodes apart of the focused group
        this.focusedEdges = []; // nedges that are apart of the focused group

        var ws = new WebSocket("ws://localhost:7071/updates");

        ws.onmessage = function(m) {
            this.update(JSON.parse(m.data));
        }.bind(this)

        ws.onopen = function() {
            ws.send('list');
        }
    }

    app.CoreModel.prototype.subscribe = function(onChange) {
        this.onChanges.push(onChange);
    }

    app.CoreModel.prototype.inform = function() {
        this.onChanges.forEach(function(cb) {
            cb();
        });
    }

    app.Entity = function() {}

    app.Entity.prototype.setPosition = function(p) {
        this.data.position.x = p.x;
        this.data.position.y = p.y;
        this.model.inform();
        dm.push(this.id, function() {
            app.Utils.request(
                "PUT",
                this.instance() + "s/" + this.data.id + "/position", // would be nice to change API to not have the "S" in it!
                p,
                null
            );
        }.bind(this), 50)
    }

    app.Group = function(data, model) {
        this.data = data;
        this.model = model;
    }

    app.Group.prototype = new app.Entity();

    app.Group.prototype.instance = function() {
        return "group";
    }

    app.Group.prototype.refreshFocusedGroup = function() {
        var model = this.model;
        var id = this.data.id;

        model.focusedNodes = model.entities[id].data.children.map(function(id) {
            return this.entities[id];
        }.bind(model))

        model.focusedEdges = model.edges.filter(function(e) {
            switch (e.instance()) {
                case 'connection':
                    if (this.entities[id].data.children.indexOf(e.data.to.id) !== -1) {
                        return true;
                    }
                    break;
                case 'link':
                    if (this.entities[id].data.children.indexOf(e.data.block.id) !== -1) {
                        return true;
                    }
                    break;
            }
            return false;
        }.bind(model))
    }

    /* setFocusedGroup sets takes a group id and prepares that group to be 
     * viewed. It changes the model's current group in focus, in addition to
     * preparing focusedNodes and focusedEdges.
     */
    app.Group.prototype.setFocusedGroup = function() {
        this.model.focusedGroup = this.data.id;
        this.refreshFocusedGroup();
        this.model.inform();
    }

    app.Block = function(data, model) {
        this.data = data;
        this.model = model;
    }

    app.Block.prototype = new app.Entity();

    app.Block.prototype.instance = function() {
        return "block";
    }

    app.Source = function(data, model) {
        this.data = data;
        this.model = model;
    }

    app.Source.prototype = new app.Entity();

    app.Source.prototype.instance = function() {
        return "source";
    }

    app.Connection = function(data, model) {
        this.data = data;
        this.model = model;
    }

    app.Connection.prototype = new app.Entity();

    app.Connection.prototype.instance = function() {
        return "connection";
    }

    app.Link = function(data, model) {
        this.data = data;
        this.model = model;
    }

    app.Link.prototype = new app.Entity();

    app.Link.prototype.instance = function() {
        return "link";
    }

    var nodes = {
        'block': app.Block,
        'source': app.Source,
        'group': app.Group,
        'connection': app.Connection,
        'link': app.Link
    }

    // this takes an id and puts it at the very top of the list
    app.CoreModel.prototype.select = function(id) {
        this.list.push(this.list.splice(this.list.indexOf(this.entities[id]), 1)[0]);
        this.inform();
    }

    app.CoreModel.prototype.addChild = function(group, id) {
        this.entities[group].data.children.push(id);
        if (group === this.focusedGroup) this.entities[group].refreshFocusedGroup();
        this.inform();
    }

    app.CoreModel.prototype.removeChild = function(group, id) {
        this.entities[group].data.children.splice(this.entities[group].data.children.indexOf(id), 1);
        if (group === this.focusedGroup) this.entities[group].refreshFocusedGroup();
        this.inform();
    }

    app.CoreModel.prototype.update = function(m) {
        switch (m.action) {
            case 'update':
                for (var key in m.data[m.type]) {
                    if (key !== 'id') {
                        this.entities[m.data[m.type].id][key] = m.data[m.type][key]
                    }
                }
                break;
            case 'create':
                // create seperate action for child.
                if (m.type === "child") {
                    this.addChild(m.data.group.id, m.data.child.id);
                    return;
                }

                // we put a reference to model in each entitiy so that we can
                // propagate inform();
                var n = new nodes[m.type](m.data[m.type], this);
                this.entities[m.data[m.type].id] = n;
                this.list.push(this.entities[m.data[m.type].id]);

                if (m.type === "group") {
                    this.groups.push(n);
                }

                if (m.type === "connection" || m.type === "link") {
                    this.edges.push(n);
                }

                break;
            case 'delete':
                if (m.type === "child") {
                    this.removeChild(m.data.group.id, m.data.child.id); // this child nonsense is a mess
                    return
                }

                var i = this.list.indexOf(this.entities[m.data[m.type].id]);
                this.list.splice(i, 1);

                if (m.type === "group") {
                    var i = this.groups.indexOf(this.entities[m.data[m.type].id]);
                    this.groups.splice(i, 1);
                }

                if (m.type === "connection" || m.type == "link") {
                    var i = this.edges.indexOf(this.entities[m.data[m.type].id]);
                    this.edges.splice(i, 1);
                }

                delete this.entities[m.data[m.type].id];
                break;
        }

        this.inform();
    }
})();
