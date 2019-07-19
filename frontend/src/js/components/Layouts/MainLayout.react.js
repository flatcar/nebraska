import PropTypes from 'prop-types';
import React from "react"
import { Row, Col } from "react-bootstrap"
import ApplicationsList from "../Applications/List.react"
import ActivityContainer from "../Activity/Container.react"

class MainLayout extends React.Component {

  constructor() {
    super()
  }

  render() {
    return(
      <div className="container">
        <Row>
          <ApplicationsList />
          <ActivityContainer />
        </Row>
      </div>
    )
  }

}

MainLayout.propTypes = {
  stores: PropTypes.object.isRequired
}

export default MainLayout
