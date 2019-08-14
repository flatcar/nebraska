import { activityStore } from "../../stores/Stores"
import React from "react"
import List from "./List.react"
import _ from "underscore"
import Loader from "react-spinners/ScaleLoader"
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';

class Container extends React.Component {

  constructor() {
    super()
    this.onChange = this.onChange.bind(this);

    this.state = {entries: activityStore.getCachedActivity()}
  }

  componentDidMount() {
    activityStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    activityStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      entries: activityStore.getCachedActivity()
    })
  }

  render() {
    let entries = ""

    if (_.isNull(this.state.entries)) {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    } else {
      if (_.isEmpty(this.state.entries)) {
        entries = <div className="emptyBox">No activity found for the last week.<br/><br/>You will see here important events related to the rollout of your updates. Stay tuned!</div>
      } else {
        entries = Object.values(_.mapObject(this.state.entries, (entry, key) => {
          return <List day={key} entries={entry} key={key} />
        }));
      }
    }

    return(
      <Grid
        container
        direction="column"
        alignItems="flex-start"
        justify="flex-start"
        className="timeline--container">
        <Grid item>
          <Typography variant="h4" className="displayInline mainTitle padBottom25">Activity</Typography>
        </Grid>
        <Grid item>
          {entries}
        </Grid>
      </Grid>
    )
  }

}

export default Container
