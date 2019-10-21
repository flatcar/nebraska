import Box from '@material-ui/core/Box';
import Paper from '@material-ui/core/Paper';
import React from "react";
import _ from "underscore";
import { activityStore } from "../../stores/Stores";
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import List from "./List.react";

function Container(props) {
  const [activity, setActivity] = React.useState(activityStore.getCachedActivity());

  React.useEffect(() => {
    activityStore.addChangeListener(onChange);

    return function cleanup () {
      activityStore.removeChangeListener(onChange);
    }
  },
  [activity]);

  function onChange() {
    setActivity(activityStore.getCachedActivity());
  }

  return(
    <Paper>
      <ListHeader title="Activity" />
      <Box padding="1em">
        { _.isNull(activity) ?
          <Loader />
        : _.isEmpty(activity) ?
          <Empty>
            No activity found for the last week.
            <br/><br/>
            You will see here important events related to the rollout of your updates. Stay tuned!
          </Empty>
        :
          Object.keys(activity).map(key => {
            const entry = activity[key];
            return (
              <List day={key} entries={entry} key={key} />
            );
          })
        }
      </Box>
    </Paper>
  );
}

export default Container;
