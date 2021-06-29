import { Box, Divider, Typography } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import ScheduleIcon from '@material-ui/icons/Schedule';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Application } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { CardFeatureLabel, CardHeader, CardLabel } from '../Common/Card';
import ListItem from '../Common/ListItem';
import MoreMenu from '../Common/MoreMenu';
import GroupsList from './ApplicationItemGroupsList';

const useStyles = makeStyles(theme => ({
  root: {
    padding: '0px 8px',
  },
  itemSection: {
    padding: '0 1em',
  },
}));

function Item(props: {
  application: Application;
  handleUpdateApplication: (appID: string) => void;
}) {
  const classes = useStyles();
  const { t } = useTranslation();
  const description = props.application.description || t('applications|No description provided');
  const groups = props.application.groups || [];
  const instances = props.application.instances.count || t('applications|None');
  const appID = props.application ? props.application.id : '';

  function updateApplication() {
    props.handleUpdateApplication(props.application.id);
  }

  function deleteApplication() {
    const confirmationText = t('applications|Are you sure you want to delete this application?');
    if (window.confirm(confirmationText)) {
      applicationsStore.deleteApplication(props.application.id);
    }
  }

  return (
    <ListItem className={classes.root}>
      <Grid container>
        <Grid item xs={12}>
          <CardHeader
            cardMainLinkLabel={props.application.name}
            cardMainLinkPath={{ pathname: `/apps/${appID}` }}
            cardId={appID}
            cardTrack=""
            cardDescription={description}
          >
            <MoreMenu
              options={[
                {
                  label: t('frequent|Edit'),
                  action: updateApplication,
                },
                {
                  label: t('frequent|Delete'),
                  action: deleteApplication,
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid item xs={12}>
          <Grid container className={classes.itemSection} spacing={0}>
            <Grid item xs={4}>
              <Box mt={2}>
                <CardFeatureLabel>{t('applications|INSTANCES')}</CardFeatureLabel>
                <CardLabel>
                  <Typography variant="h5">{instances}</Typography>
                  <Box display="flex" my={1}>
                    <ScheduleIcon color="disabled" />
                    <Box pl={1} color="text.disabled">
                      <Typography variant="subtitle1">{t('applications|last 24 hours')}</Typography>
                    </Box>
                  </Box>
                </CardLabel>
              </Box>
            </Grid>
            <Box width="1%">
              <Divider orientation="vertical" variant="fullWidth" />
            </Box>
            <Grid item xs={7}>
              <Box mt={2} p={1}>
                <CardFeatureLabel>{t('frequent|Groups')}</CardFeatureLabel>
                <Box display="inline-block" pl={2}>
                  <CardLabel>{groups.length === 0 ? t('applications|None') : groups.length}</CardLabel>
                </Box>
                <GroupsList
                  groups={groups}
                  appID={props.application.id}
                  appName={props.application.name}
                />
              </Box>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </ListItem>
  );
}

export default Item;
