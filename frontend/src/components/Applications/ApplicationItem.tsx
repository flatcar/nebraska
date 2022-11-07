import { Box, Divider, Typography } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import ScheduleIcon from '@material-ui/icons/Schedule';
import { useTranslation } from 'react-i18next';
import { Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { CardFeatureLabel, CardHeader, CardLabel } from '../common/Card/Card';
import ListItem from '../common/ListItem';
import MoreMenu from '../common/MoreMenu';
import ApplicationItemGroupsList from './ApplicationItemGroupsList';

const useStyles = makeStyles({
  root: {
    padding: '0px 8px',
  },
  itemSection: {
    padding: '0 1em',
  },
});

export interface ApplicationItemProps {
  onUpdate: (appID: string) => void;
  description?: string;
  groups: Group[] | null;
  numberOfInstances: number;
  id: string;
  productId: string;
  name: string;
}

export default function ApplicationItem(props: ApplicationItemProps) {
  const classes = useStyles();
  const { t } = useTranslation();
  const { description, groups, numberOfInstances, id, productId, name } = props;

  return (
    <ListItem className={classes.root}>
      <Grid container>
        <Grid item xs={12}>
          <CardHeader
            cardMainLinkLabel={name}
            cardMainLinkPath={{ pathname: `/apps/${id}` }}
            cardId={id}
            cardTrack={productId}
            cardDescription={description || t('applications|No description provided')}
          >
            <MoreMenu
              options={[
                {
                  label: t('frequent|Edit'),
                  action: () => props.onUpdate(id),
                },
                {
                  label: t('frequent|Delete'),
                  action: () => {
                    window.confirm(
                      t('applications|Are you sure you want to delete this application?')
                    )
                      ? applicationsStore().deleteApplication(id)
                      : null;
                  },
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
                  <Typography variant="h5">
                    {numberOfInstances || t('applications|None')}
                  </Typography>
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
                  <CardLabel>
                    {groups?.length === 0 ? t('applications|None') : groups?.length}
                  </CardLabel>
                </Box>
                <ApplicationItemGroupsList groups={groups} appID={id} appName={name} />
              </Box>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </ListItem>
  );
}
