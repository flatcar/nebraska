import Card from '@material-ui/core/Card';
import { makeStyles } from '@material-ui/core/styles';
import moment from "moment";
import PropTypes from 'prop-types';
import React from "react";
import { activityStore } from '../../stores/Stores';

const useStyles = makeStyles(theme => ({
  activityCard: {
    overflow: "unset",
  },
}));

function ActivityCard(props) {
  const classes = useStyles();
  return (
    <Card className={`${classes.activityCard} timeline--eventLabel`}>{props.children}</Card>
  );
}

class Item extends React.Component {

  constructor(props) {
    super(props)

    this.state = {
      entryClass: {},
      entrySeverity: {}
    }
  }

  fetchEntryClassFromStore() {
    let entryClass = activityStore.getActivityEntryClass(this.props.entry.class, this.props.entry)
    this.setState({
      entryClass: entryClass
    })
  }

  fetchEntrySeverityFromStore() {
    let entrySeverity = activityStore.getActivityEntrySeverity(this.props.entry.severity)
    this.setState({
      entrySeverity: entrySeverity
    })
  }

  componentDidMount() {
    this.fetchEntryClassFromStore()
    this.fetchEntrySeverityFromStore()
  }

  render() {
    let ampm = moment.utc(this.props.entry.created_ts).local().format("a"),
        time = moment.utc(this.props.entry.created_ts).local().format("HH:mm"),
        subtitle = "",
        name = ""

    if (this.state.entryClass.type !== "activityChannelPackageUpdated") {
      subtitle = "GROUP:"
      name = this.state.entryClass.groupName
    }

    return (
      <li className = {this.state.entrySeverity.className}>
        <div className="timeline--icon">
          <span className={"fa " + this.state.entrySeverity.icon}></span>
        </div>
        <div className="timeline--event">
          {time}
          <br />
          <span className="timeline--ampm">{ampm}</span>
        </div>
        <ActivityCard>
          <div className="row timeline--eventLabelTitle">
            <div className="col-xs-5 noPadding">{this.state.entryClass.appName}</div>
            <div className="col-xs-7 noPadding">
              <span className="subtitle">{subtitle} </span>
              {name}
            </div>
          </div>
          <p>{this.state.entryClass.description}</p>
        </ActivityCard>
      </li>
    )
  }

}

Item.propTypes = {
  entry: PropTypes.object.isRequired
};

export default Item
