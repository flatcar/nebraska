import { activityStore } from "../../stores/Stores"
import React from "react"
import List from "./List.react"
import _ from "underscore"
import Loader from '../Common/Loader';
import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import ListHeader from '../Common/ListHeader';
import Empty from '../Common/EmptyContent';

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
      entries = <Loader />
    } else {
      if (_.isEmpty(this.state.entries)) {
        entries = <Empty>No activity found for the last week.<br/><br/>You will see here important events related to the rollout of your updates. Stay tuned!</Empty>
      } else {
        entries = Object.values(_.mapObject(this.state.entries, (entry, key) => {
          return <List day={key} entries={entry} key={key} />
        }));
      }
    }

    return(
      <Paper>
        <ListHeader title="Activity" />
        <Box padding="1em">
          {entries}
        </Box>
      </Paper>
    )
  }

}

export default Container
