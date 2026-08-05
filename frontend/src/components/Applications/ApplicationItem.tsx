import ScheduleIcon from '@mui/icons-material/Schedule';
import { Box, Divider, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { styled } from '@mui/material/styles';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { Group } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { CardFeatureLabel, CardHeader, CardLabel } from '../common/Card/Card';
import ConfirmDialog from '../common/ConfirmDialog';
import ListItem from '../common/ListItem';
import MoreMenu from '../common/MoreMenu';
import ApplicationItemGroupsList from './ApplicationItemGroupsList';

const PREFIX = 'ApplicationItem';

const classes = {
  root: `${PREFIX}-root`,
  itemSection: `${PREFIX}-itemSection`,
  instancesPanel: `${PREFIX}-instancesPanel`,
};

const StyledListItem = styled(ListItem)(({ theme }) => ({
  [`&.${classes.root}`]: {
    padding: '0px 8px',
  },
  [`& .${classes.itemSection}`]: {
    padding: '8px 1em 1em',
  },
  [`& .${classes.instancesPanel}`]: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
    gap: theme.spacing(0.5),
    padding: theme.spacing(1.5),
    borderRadius: 12,
    backgroundColor: 'rgba(11, 124, 133, 0.04)',
    border: '1px solid rgba(11, 124, 133, 0.1)',
  },
}));

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
  const [confirmDeleteOpen, setConfirmDeleteOpen] = React.useState(false);

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
                  action: () => setConfirmDeleteOpen(true),
                },
              ]}
            />
          </CardHeader>
        </Grid>
        <Grid size={12}>
          <Grid
            container
            className={classes.itemSection}
            columnSpacing={2}
            rowSpacing={0}
            alignItems="flex-start"
          >
            <Grid size={4}>
              <Box className={classes.instancesPanel}>
                <CardFeatureLabel>{t('applications|instances_title')}</CardFeatureLabel>
                <Typography
                  variant="h4"
                  sx={{
                    fontWeight: 800,
                    color: numberOfInstances ? 'text.primary' : 'text.secondary',
                  }}
                >
                  {numberOfInstances || t('applications|none')}
                </Typography>
                <Box display="flex" alignItems="center" color="text.secondary">
                  <ScheduleIcon sx={{ fontSize: 16, mr: 0.75 }} color="disabled" />
                  <Typography variant="body2" color="text.secondary">
                    {t('applications|time_last_24_hours')}
                  </Typography>
                </Box>
              </Box>
            </Grid>
            <Box width="1%">
              <Divider orientation="vertical" variant="fullWidth" />
            </Box>
            <Grid size={7}>
              <Box px={1} pt={0}>
                <Box display="flex" alignItems="baseline" gap={1} mb={1}>
                  <CardFeatureLabel>{t('frequent|groups')}</CardFeatureLabel>
                  <CardLabel
                    labelStyle={{
                      fontWeight: 700,
                      color: '#0B7C85',
                    }}
                  >
                    {groups?.length === 0 ? t('applications|none') : groups?.length}
                  </CardLabel>
                </Box>
                <ApplicationItemGroupsList groups={groups} appID={id} appName={name} />
              </Box>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
      <ConfirmDialog
        open={confirmDeleteOpen}
        description={t('applications|confirm_delete_application')}
        onClose={() => setConfirmDeleteOpen(false)}
        onConfirm={() => applicationsStore().deleteApplication(id)}
      />
    </StyledListItem>
  );
}
