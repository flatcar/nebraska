import ScheduleIcon from '@mui/icons-material/Schedule';
import { Box, Divider, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { styled } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';

import { Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { CardFeatureLabel, CardHeader, CardLabel } from '../common/Card/Card';
import ListItem from '../common/ListItem';
import MoreMenu from '../common/MoreMenu';
import ApplicationItemGroupsList from './ApplicationItemGroupsList';

const PREFIX = 'ApplicationItem';

const classes = {
  root: `${PREFIX}-root`,
  itemSection: `${PREFIX}-itemSection`,
};

const StyledListItem = styled(ListItem)({
  [`&.${classes.root}`]: {
    padding: '0px 8px',
  },
  [`& .${classes.itemSection}`]: {
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
  const { t } = useTranslation();
  const { description, groups, numberOfInstances, id, productId, name } = props;

  return (
    <StyledListItem className={classes.root}>
      <Grid container sx={{ width: '100%' }}>
        <Grid size={12}>
          <CardHeader
            cardMainLinkLabel={name}
            cardMainLinkPath={{ pathname: `/apps/${id}` }}
            cardId={id}
            cardTrack={productId}
            cardDescription={description || t('applications|description_none_provided')}
          >
            <MoreMenu
              options={[
                {
                  label: t('frequent|edit'),
                  action: () => props.onUpdate(id),
                },
                {
                  label: t('frequent|delete'),
                  action: () => {
                    if (window.confirm(t('applications|confirm_delete_application'))) {
                      applicationsStore().deleteApplication(id);
                    }
                  },
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid size={12}>
          <Grid container className={classes.itemSection} spacing={0}>
            <Grid size={4}>
              <Box mt={2}>
                <CardFeatureLabel>{t('applications|instances_title')}</CardFeatureLabel>
                <CardLabel>
                  <Typography variant="h5">
                    {numberOfInstances || t('applications|none')}
                  </Typography>
                  <Box display="flex" my={1}>
                    <ScheduleIcon color="disabled" />
                    <Box pl={1} color="text.disabled">
                      <Typography variant="subtitle1">
                        {t('applications|time_last_24_hours')}
                      </Typography>
                    </Box>
                  </Box>
                </CardLabel>
              </Box>
            </Grid>
            <Box width="1%">
              <Divider orientation="vertical" variant="fullWidth" />
            </Box>
            <Grid size={7}>
              <Box mt={2} p={1}>
                <CardFeatureLabel>{t('frequent|groups')}</CardFeatureLabel>
                <Box display="inline-block" pl={2}>
                  <CardLabel>
                    {groups?.length === 0 ? t('applications|none') : groups?.length}
                  </CardLabel>
                </Box>
                <ApplicationItemGroupsList groups={groups} appID={id} appName={name} />
              </Box>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </StyledListItem>
  );
}
