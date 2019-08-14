import Grid from '@material-ui/core/Grid';
import PropTypes from 'prop-types';
import React from 'react';
import Item from './Item.react';

class List extends React.Component {

  constructor() {
    super()
  }

  render() {
    let entries = this.props.entries ? this.props.entries : []

    return(
      <Grid
        container
        alignItems="stretch"
        direction="column">
      <h5 className="timeline--contentTitle">
        {this.props.day}
      </h5>
        <ul className="timeline--content">
          {entries.map((entry, i) =>
            <Item key={i} entry={entry} />
          )}
        </ul>
      </Grid>
    )
  }

}

List.propTypes = {
  day: PropTypes.string.isRequired,
  entries: PropTypes.array.isRequired
}

export default List
