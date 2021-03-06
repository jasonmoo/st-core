var app = app || {};

/* AutoCompleteComponenet
 * Given a list of items and a position by parent.
 * Fires onChange event to parent node when item is selected.
 */

(function() {
    'use strict';

    app.AutoCompleteComponent = React.createClass({
        displayName: 'AutoCompleteComponent',
        getInitialState: function() {
            return {
                text: '',
                options: [],
            }
        },
        componentWillMount: function() {
            this.autocomplete(this.state.text);
        },
        componentDidMount: function() {
          this.getDOMNode().children[0].focus()
        },
        autocomplete: function(s) {
            var reBegin = new RegExp('^' + app.Utils.escape(s), 'i');
            var reIn = new RegExp(app.Utils.escape(s), 'i');

            var beginList = this.props.options.filter(function(o) {
                return o.type.match(reBegin)
            }).sort(function(a, b) {
                if (a.type > b.type) return 1;
                if (a.type < b.type) return -1;
                return 0
            });

            var inList = this.props.options.filter(function(o) {
                return o.type.match(reIn);
            }).sort(function(a, b) {
                if (a.type > b.type) return 1;
                if (a.type < b.type) return -1;
                return 0;
            })

            this.setState({
                options: beginList.concat(inList.filter(function(o) {
                    return beginList.indexOf(o) === -1;
                }))
            });
        },
        handleChange: function(e) {
            this.autocomplete(e.target.value);
        },
        handleKeyUp: function(e) {
            if (e.nativeEvent.keyCode === 13) {
                var match = this.props.options.filter(function(o) {
                    return o.type === e.target.value
                });
                if (match.length === 1) {
                    this.props.onChange(match[0]);
                }
            }
        },
        render: function() {
            var input = React.createElement('input', {
                onChange: this.handleChange,
                onKeyUp: this.handleKeyUp,
                key: 'autocomplete_input'
            }, null);

            var options = this.state.options.map(function(o) {
                return React.createElement('li', {
                    key: o.type
                }, o.type);
            });

            var list = React.createElement('ul', {
                key: 'autocomplete_list',
            }, options);

            return React.createElement('div', {
                style: {
                    width: '100',
                    top: this.props.y,
                    left: this.props.x,
                    position: 'absolute',
                    zIndex: 100,
                },
            }, [input, list]);
        }
    });
})();
