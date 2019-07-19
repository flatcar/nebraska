import PropTypes from 'prop-types';
import React from "react"
import { Row, Col } from "react-bootstrap"
import Item from "./Item.react"

class List extends React.Component {

  constructor() {
    super()
  }

  render() {
    let entries = this.props.entries ? this.props.entries : []

    return(
      <div>
        <h5 className="timeline--contentTitle">
          {this.props.day}
        </h5>
        <Row>
          <ul className="timeline--content">
            {entries.map((entry, i) =>
              <Item key={i} entry={entry} />
            )}
          </ul>
        </Row>
      </div>
    )
  }

}

List.propTypes = {
  day: PropTypes.string.isRequired,
  entries: PropTypes.array.isRequired
}

export default List
