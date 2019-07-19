import PropTypes from 'prop-types';
import React from "react"
import ReactDOM from "react-dom"
import PubSub from "pubsub-js"
import ProgressBarJS from "progressbar.js"

class ProgressBar extends React.Component {

  constructor(props) {
    super(props)

    this.line = null
    this.inProgress = false
    this.inProgressCount = 0
  }

  componentDidMount() {
    let lineContainer = ReactDOM.findDOMNode(this.refs.progressBar)
    let lineOptions = {
      color: this.props.color,
      strokeWidth: this.props.width,
      easing: "easeInOut"
    }
    this.line = new ProgressBarJS.Line(ReactDOM.findDOMNode(lineContainer), lineOptions)
  }

  componentWillMount() {
    PubSub.subscribe(this.props.name, (t, m) => { return this.handleMsg(m) })
  }

  componentWillUnmount() {
    PubSub.unsubscribe(this.props.name)
  }

  render() {
    return React.createElement("div", {
      className: this.props.containerClassName,
      ref: "progressBar"
    })
  }

  handleMsg(msg) {
    if (this.inProgress) {
      switch (msg) {
        case "add":
          this.inProgressCount++
          break
        case "done":
          if (this.inProgressCount > 0) {
            this.inProgressCount--
          }
          if (this.inProgressCount == 0) {
            this.line.animate(1.0, {duration: 200}, () => {
              this.line.set(0)
              this.inProgress = false
            })
          }
          break
      }
    } else {
      switch (msg) {
        case "add":
          this.inProgressCount++
          this.inProgress = true
          this.line.animate(0.25, {duration: 5000}, () => {
            this.line.animate(0.75, {duration: 10000})
          })
          break
      }
    }
  }

};

ProgressBar.propTypes = {
  name: PropTypes.string.isRequired,
  color: PropTypes.string.isRequired,
  width: PropTypes.number.isRequired
}

export default ProgressBar
