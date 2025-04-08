import { FormControl, Grid, InputLabel, MenuItem, Select, SelectChangeEvent } from '@mui/material';
import { Field } from 'formik';
import { TextField } from 'formik-mui';
import { useTranslation } from 'react-i18next';

import { Channel } from '../../../api/apiDataTypes';
import { ARCHES } from '../../../utils/helpers';

export interface GroupDetailsFormProps {
  channels: Channel[];
  values: { [key: string]: string };
  setFieldValue: (formField: string, value: any) => any;
}

export default function GroupDetailsForm(props: GroupDetailsFormProps) {
  const { t } = useTranslation();
  const { channels, values, setFieldValue } = props;

  return (
    <div style={{ padding: '1rem' }}>
      <Grid container spacing={2} justifyContent="center">
        <Grid item xs={8}>
          <Field
            name="name"
            component={TextField}
            variant="standard"
            margin="dense"
            label="Name"
            required
            fullWidth
            value={values.name}
            onChange={(e: any) => {
              setFieldValue('name', e.target.value);
            }}
          />
        </Grid>
        <Grid item xs={4}>
          <FormControl margin="dense" fullWidth>
            <InputLabel variant="standard" shrink>
              {t('groups|channel')}
            </InputLabel>
            <Field
              name="channel"
              component={Select}
              variant="standard"
              displayEmpty
              defaultValue={values.channel}
              onChange={(e: SelectChangeEvent) => {
                setFieldValue('channel', e.target.value);
              }}
            >
              <MenuItem value="" key="">
                {t('groups|none_yet')}
              </MenuItem>
              {channels.map((channelItem: Channel) => (
                <MenuItem value={channelItem.id} key={channelItem.id}>
                  {`${channelItem.name}(${ARCHES[channelItem.arch]})`}
                </MenuItem>
              ))}
            </Field>
          </FormControl>
        </Grid>
        <Grid item xs={12}>
          <Field
            name="track"
            component={TextField}
            variant="standard"
            margin="dense"
            label={t('groups|track_identifier')}
            fullWidth
            value={values.track}
            onChange={(e: any) => {
              setFieldValue('track', e.target.value);
            }}
          />
        </Grid>
        <Grid item xs={12}>
          <Field
            name="description"
            component={TextField}
            variant="standard"
            margin="dense"
            label={t('groups|description')}
            fullWidth
            value={values.description}
            onChange={(e: any) => {
              setFieldValue('description', e.target.value);
            }}
          />
        </Grid>
      </Grid>
    </div>
  );
}
