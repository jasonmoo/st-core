var app = app || {};

/* DragContainerComponent
 * This component wraps around an element of model Entity and provides
 * mouse-drag functionality.
 */

(function() {
    'use strict';

    app.DragContainer = React.createClass({
        displayName: 'DragContainer',
        getInitialState: function() {
            return {
                dragging: false,
                offX: null,
                offY: null,
                debounce: 0,
            }
        },
        onMouseDown: function(e) {
            this.props.nodeSelect(this.props.model.data.id);

            // see TODO in model.js in CoreModel.update()
            // this is to protect our model from receiving updates
            // from the server
            this.props.model.setDragging(true);

            this.setState({
                dragging: true,
                offX: e.pageX - this.props.x,
                offY: e.pageY - this.props.y
            })
        },
        componentDidUpdate: function(props, state) {
            if (this.state.dragging && !state.dragging) {
                document.addEventListener('mousemove', this.onMouseMove)
                document.addEventListener('mouseup', this.onMouseUp)
            } else if (!this.state.dragging && state.dragging) {
                document.removeEventListener('mousemove', this.onMouseMove)
                document.removeEventListener('mouseup', this.onMouseUp)
            }
        },
        onMouseUp: function(e) {
            this.props.model.setPosition({
                x: e.pageX - this.state.offX,
                y: e.pageY - this.state.offY
            });

            this.setState({
                dragging: false,
            });
            // see TODO in model.js in CoreModel.update()
            // this is to protect our model from receiving updates
            // would like to do away with this...
            this.props.model.setDragging(false);
            this.props.onDragStop();
        },
        onMouseMove: function(e) {
            if (this.state.dragging) {
                var diffX = e.pageX - this.state.offX;
                var diffY = e.pageY - this.state.offY;

                this.props.onDrag(-1 * (this.props.model.data.position.x - diffX), -1 * (this.props.model.data.position.y - diffY));
            }
        },
        render: function() {
            return (
                React.createElement('g', {
                        transform: 'translate(' + this.props.x + ', ' + this.props.y + ')',
                        onMouseMove: this.onMouseMove,
                        onMouseDown: this.onMouseDown,
                        onMouseUp: this.onMouseUp,
                    },
                    this.props.children
                )
            )

        }
    })
})();
