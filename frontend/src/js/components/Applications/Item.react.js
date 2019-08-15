import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import { Link as RouterLink } from "react-router-dom"
import GroupsList from "./ApplicationItemGroupsList.react"
import ChannelsList from "./ApplicationItemChannelsList.react"
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Link from '@material-ui/core/Link';
import Typography from '@material-ui/core/Typography';
import { styled } from '@material-ui/styles';
import { makeStyles } from '@material-ui/core/styles';

const GridHeader = styled(Grid)({
  background: '#fafafa',
  padding: 16,
});

const AppLink = styled(Link)({
  color: 'rgba(119,119,119)',
  fontSize: 18,
});

const AppIdLabel = styled(Typography)({
  color: 'rgba(119,119,119,.75)',
  fontSize: 12
});

const AppDescriptionLabel = styled(Typography)({
  color: 'rgb(119,119,119, .9)',
  fontSize: 14,
});

const useStyles = makeStyles(theme => ({
  featureLabel: {
    color: 'rgb(119,119,119, .9)',
    fontVariant: 'small-caps',
    fontSize: 14,
  },
}));

function CardFeatureLabel(props) {
  const classes = useStyles();
  return (
    <Typography component='span' className={classes.featureLabel}>{props.children}</Typography>
  );
}

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.updateApplication = this.updateApplication.bind(this)
    this.deleteApplication = this.deleteApplication.bind(this)
  }

  updateApplication() {
    this.props.handleUpdateApplication(this.props.application.id)
  }

  deleteApplication() {
    let confirmationText = "Are you sure you want to delete this application?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteApplication(this.props.application.id)
    }
  }

  render() {
    let application = this.props.application ? this.props.application : {},
        description = this.props.application.description ? this.props.application.description : "No description provided",
        styleDescription = this.props.application.description ? "" : " italicText",
        channels = this.props.application.channels ? this.props.application.channels : [],
        groups = this.props.application.groups ? this.props.application.groups : [],
        instances = this.props.application.instances.count ? this.props.application.instances.count : 0,
        appID = this.props.application ? this.props.application.id : "",
        popoverContent = {
          type: "application",
          appID: appID
        }

    return(
      <Card>
        <GridHeader container
          direction="column">
          <Grid container>
            <Grid item xs={5}>
              <AppLink component={RouterLink} to={{pathname: `/apps/${appID}`}}>
                {this.props.application.name} <i className="fa fa-caret-right"></i>
              </AppLink>
            </Grid>
            <Grid item xs={5}>
              <AppIdLabel>ID: {appID}</AppIdLabel>
            </Grid>
            <Grid item xs={2}>
              <IconButton aria-label="delete" disabled color="primary">
                <DeleteIcon />
              </IconButton>
              <div className="apps--buttons">
                <button className="cr-button displayInline fa fa-edit" onClick={this.updateApplication}></button>
                <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteApplication}></button>
              </div>
            </Grid>
          </Grid>
          <Grid item>
            <AppDescriptionLabel variant="h5">{description}</AppDescriptionLabel>
          </Grid>
        </GridHeader>
        <CardContent>
          <Grid container spacing={2}>
            <Grid item xs={2}>
              <CardFeatureLabel>Instances:</CardFeatureLabel>
              <span>
                {instances}
              </span>
            </Grid>
            <Grid item xs={10}>
              <CardFeatureLabel>Groups:</CardFeatureLabel>
              <span>
                {groups.length}
              </span>
              <GroupsList
                groups={groups}
                appID={this.props.application.id}
                appName={this.props.application.name} />
            </Grid>
            <Grid item xs={12}>
              <CardFeatureLabel>Channels:</CardFeatureLabel>
              <ChannelsList channels={channels} />
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    )
  }
}

Item.propTypes = {
  application: PropTypes.object.isRequired,
  handleUpdateApplication: PropTypes.func.isRequired
}

export default Item
