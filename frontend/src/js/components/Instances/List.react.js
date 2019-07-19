import PropTypes from 'prop-types';
import React from "react"
import Item from "./Item.react"
import _ from "underscore"
import semver from "semver"
import { cleanSemverVersion } from "../../constants/helpers"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onItemToggle = this.onItemToggle.bind(this)

    this.state = {selections: {}}
  }

  onItemToggle(id, selected) {
    let selections = this.state.selections
    if (selected == true) {
      _.mapObject(selections, (val, key) => {
        if (val == true) {
          selections[key] = false
        }
      })
      // Save opened instance
      this.props.onChangeSelectedInstance(id)
    } else {
      // Remove opened instance
      this.props.onChangeSelectedInstance("")
    }

    selections[id] = selected;
    this.setState({
      selections: selections
    })
  }

  render() {
    let versions = this.props.version_breakdown ? this.props.version_breakdown : [],
        lastVersionChannel = this.props.channel.package ? cleanSemverVersion(this.props.channel.package.version) : "",
        versionNumbers = (_.map(versions, (version) => {
          return cleanSemverVersion(version.version)
        })).sort(semver.rcompare)

    return(
      <div className="coreRollerTable">
        <div className="coreRollerTable-header">
          <div className="coreRollerTable-cell">IP</div>
          <div className="coreRollerTable-cell coreRollerTable-cell--medium">Instance ID</div>
          <div className="coreRollerTable-cell">Current status</div>
          <div className="coreRollerTable-cell">Version</div>
          <div className="coreRollerTable-cell">Last check</div>
        </div>
        {this.props.instances.map((instance, i) =>
          <Item key={"instance_" + i} instance={instance} lastVersionChannel={lastVersionChannel} versionNumbers={versionNumbers} selected={this.state.selections[instance.id]} onToggle={this.onItemToggle} />
        )}
      </div>
    )
  }

}

List.propTypes = {
  instances: PropTypes.array.isRequired,
  version_breakdown: PropTypes.array,
  channel: PropTypes.object
}

export default List
