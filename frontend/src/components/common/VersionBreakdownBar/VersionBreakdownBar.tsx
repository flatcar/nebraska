import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';
import { Theme } from '@mui/material/styles';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';
import { useTheme } from '@mui/styles';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

import { Channel } from '../../../api/apiDataTypes';
import { cleanSemverVersion, makeColorsForVersions } from '../../../utils/helpers';

const PREFIX = 'VersionProgressBar';

const classes = {
  chart: `${PREFIX}-chart`,
  container: `${PREFIX}-container`
};

const BorderlessTableCell = styled(TableCell)(() => ({
  border: 'none'
}));

const StyledResponsiveContainer = styled(ResponsiveContainer)(({ theme }) => ({
  [`& .${classes.chart}`]: {
    zIndex: theme.zIndex.drawer,
  },

  [`&.${classes.container}`]: {
    marginLeft: 'auto',
    marginRight: 'auto',
  }
}));

function VersionsTooltip(props: {
  versionsData: {
    data: any;
    versions: string[];
    colors: {
      [key: string]: string;
    };
  };
}) {

  const { data, versions, colors } = props.versionsData;

  return (
    <div className="custom-tooltip">
      <Paper>
        <Table>
          <TableBody>
            {versions.map(version => {
              const color = colors[version];
              const value = data[version].toFixed(1);
              return (
                <TableRow key={version}>
                  <BorderlessTableCell>
                    <span style={{ color: color, fontWeight: 'bold' }}>{version}</span>
                  </BorderlessTableCell>
                  <BorderlessTableCell>{value} %</BorderlessTableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Paper>
    </div>
  );
}

function VersionProgressBar(props: { version_breakdown: any; channel: Channel | null }) {
  const theme = useTheme();
  const { t } = useTranslation();
  let lastVersionChannel: string | null = '';
  const otherVersionLabel = t('common|other_option');
  const [chartData, setChartData] = React.useState<{
    data: any;
    versions: string[];
    colors: {
      [key: string]: string;
    };
  }>({
    data: {},
    versions: [],
    colors: {},
  });

  function setup(version_breakdown: any, channel: Channel | null) {
    const data: { [key: string]: any } = {};
    const other = {
      versions: [],
      percentage: 0,
    };

    version_breakdown.forEach((entry: never) => {
      const { version, percentage } = entry;
      const percentageValue = parseFloat(percentage);

      if (percentage < 10) {
        other.versions.push(version);
        other.percentage += percentageValue;
        return;
      }

      data[version] = percentageValue;
    });

    const versionColors = makeColorsForVersions(theme as Theme, Object.keys(data), channel);
    lastVersionChannel = channel && channel.package ? channel.package.version : null;

    if (other.percentage > 0) {
      data[otherVersionLabel] = other.percentage;
      versionColors[otherVersionLabel] = (theme as Theme).palette.grey['500'];
    }

    const versionsSorted = Object.keys(data).sort((version1, version2) => {
      // If the version is the channel's one, then it should come first.
      // If it's the 'Other', then it should come last.
      // Otherwise compare the number of instances.
      const cleanVersion1 = cleanSemverVersion(version1);
      const cleanVersion2 = cleanSemverVersion(version2);
      const results: { [key: string]: number } = { cleanVersion1: -1, cleanVersion2: 1 };

      for (const version of [cleanVersion1, cleanVersion2]) {
        switch (version) {
          case lastVersionChannel:
            return results[version];
          case otherVersionLabel:
            return -results[version];
          default:
            break;
        }
      }

      return data[cleanVersion1] - data[cleanVersion2];
    });

    data['key'] = 'version_breakdown';

    setChartData({
      data: data,
      versions: versionsSorted,
      colors: versionColors,
    });
  }

  React.useEffect(() => {
    setup(props.version_breakdown, props.channel);
  }, [props.version_breakdown, props.channel]);

  return (
    <StyledResponsiveContainer width="95%" height={30} className={classes.container}>
      <BarChart layout="vertical" maxBarSize={10} data={[chartData.data]} className={classes.chart}>
        <Tooltip content={<VersionsTooltip versionsData={chartData} />} />
        <XAxis hide type="number" />
        <YAxis hide dataKey="key" type="category" />
        {chartData.versions.map((version, index) => {
          const color = chartData.colors[version];
          return <Bar key={index} dataKey={version} stackId="1" fill={color} layout="vertical" />;
        })}
      </BarChart>
    </StyledResponsiveContainer>
  );
}

export default VersionProgressBar;
