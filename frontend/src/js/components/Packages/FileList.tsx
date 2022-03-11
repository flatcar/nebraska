import addIcon from '@iconify/icons-mdi/plus';
import Icon from '@iconify/react';
import Box from '@material-ui/core/Box';
import Button from '@material-ui/core/Button';
import Grid from '@material-ui/core/Grid';
import IconButton from '@material-ui/core/IconButton';
import List from '@material-ui/core/List';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { File, Package } from '../../api/apiDataTypes';
import ListItem from '../Common/ListItem';
import MoreMenu from '../Common/MoreMenu';

interface FileListItemProps {
  file: File;
  edit: boolean;
  onEditClicked?: () => void;
  onEditFinished: (file?: File) => void;
  onDeleteClicked?: () => void;
  showEditOptions?: boolean;
}

function FileListItem(props: FileListItemProps) {
  const { file, edit, onEditClicked, onEditFinished, onDeleteClicked, showEditOptions } = props;
  const { t } = useTranslation('frequent');
  const [fileToEdit, setFileToEdit] = React.useState(file);

  function onFileInfoChanged(
    infoToChange: Omit<keyof File, 'id' | 'created_ts'>,
    infoValue: string
  ) {
    const newFile = { ...fileToEdit, [infoToChange as string]: infoValue };
    setFileToEdit(newFile);
  }

  // If the file changes, we need to update the editing one.
  React.useEffect(() => {
    setFileToEdit(file);
  }, [file]);

  function hasSize() {
    return (parseInt(fileToEdit.size) || 0) > 0;
  }

  if (!edit) {
    return (
      <ListItem>
        <Grid container direction="column">
          <Grid item>
            <Typography>{fileToEdit.name}</Typography>
          </Grid>
          <Grid item container>
            <Grid item xs={5}>
              <Typography component="span" variant="body2" color="textSecondary">
                {t('frequent|Size: {{size}}', { size: hasSize() ? fileToEdit.size : '-' })}
              </Typography>
            </Grid>
            <Grid item>
              <Typography component="span" variant="body2" color="textSecondary">
                {t('frequent|Hash: {{hash}}', { hash: fileToEdit.hash || '-' })}
              </Typography>
            </Grid>
          </Grid>
        </Grid>
        <ListItemSecondaryAction>
          <MoreMenu
            iconButtonProps={{ disabled: showEditOptions }}
            options={[
              {
                label: t('frequent|Edit'),
                action: () => {
                  onEditClicked && onEditClicked();
                },
              },
              {
                label: t('frequent|Delete'),
                action: () => {
                  onDeleteClicked && onDeleteClicked();
                },
              },
            ]}
          />
        </ListItemSecondaryAction>
      </ListItem>
    );
  }

  return (
    <ListItem>
      <Grid container>
        <Grid item xs={12}>
          <Typography variant="h5">
            {!file.name ? t('packages|New File') : t('packages|Edit File')}
          </Typography>
        </Grid>
        <Grid item xs={12}>
          <TextField
            name="name"
            margin="dense"
            label={t('packages|Name')}
            type="text"
            required
            fullWidth
            value={fileToEdit.name}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
              onFileInfoChanged('name', e.target.value)
            }
          />
        </Grid>
        <Grid item container spacing={1}>
          <Grid item xs={6}>
            <TextField
              name="size"
              margin="dense"
              label={t('packages|Size')}
              type="text"
              helperText={t('packages|In bytes')}
              fullWidth
              value={fileToEdit.size}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
                onFileInfoChanged('size', e.target.value)
              }
            />
          </Grid>
          <Grid item xs={6}>
            <TextField
              name="hash"
              margin="dense"
              label={t('packages|Hash')}
              type="text"
              fullWidth
              value={fileToEdit.hash}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
                onFileInfoChanged('hash', e.target.value)
              }
            />
          </Grid>
        </Grid>
        <Grid item container>
          <Grid item>
            <Button
              onClick={() => {
                setFileToEdit(file);
                onEditFinished();
              }}
            >
              {t('frequent|Cancel')}
            </Button>
          </Grid>
          <Grid item>
            <Button onClick={() => onEditFinished(fileToEdit)}>{t('frequent|Done')}</Button>
          </Grid>
        </Grid>
      </Grid>
    </ListItem>
  );
}

interface FileListProps {
  files: Package['extra_files'];
  onFilesChanged: (files: File[]) => void;
  onEditChanged: (isEditing: boolean) => void;
}

export default function FileList(props: FileListProps) {
  const { files = [], onFilesChanged, onEditChanged } = props;
  const [editFileIndex, setEditFileIndex] = React.useState(-1);
  const [showFileAdd, setShowFileAdd] = React.useState(false);
  const { t } = useTranslation('packages');

  function onEditClicked(fileIndex: number) {
    setEditFileIndex(fileIndex);
    onEditChanged(fileIndex !== -1);
  }

  function onFileChanged(file: File, index: number) {
    const editedFiles = [...(files || [])];
    editedFiles[index] = file;
    onFilesChanged(editedFiles);
    stopEditing();
  }

  function stopEditing() {
    setEditFileIndex(-1);
    onEditChanged(false);
  }

  function addFile(newFile?: File) {
    if (!!newFile && !!newFile.name) {
      const editedFiles = [...(files || []), newFile];
      onFilesChanged(editedFiles);
    }
    setShowFileAdd(!showFileAdd);
    onEditChanged(!showFileAdd);
  }

  function isEditing() {
    return showFileAdd || editFileIndex !== -1;
  }

  function deleteFile(index: number) {
    onFilesChanged(files.filter((v, i) => i !== index));
  }

  return (
    <List style={{ minHeight: 'calc(50vh)' }}>
      {files?.map((file, i) => (
        <FileListItem
          key={`file_list_item_${i}`}
          file={file}
          showEditOptions={isEditing()}
          onEditFinished={(file?: File) => (!!file ? onFileChanged(file, i) : stopEditing())}
          edit={editFileIndex === i}
          onEditClicked={() => onEditClicked(i)}
          onDeleteClicked={() => deleteFile(i)}
        />
      ))}
      {showFileAdd ? (
        <FileListItem
          file={{ name: '', size: '', hash: '' }}
          onEditFinished={(file?: File) => addFile(file)}
          edit
        />
      ) : (
        <Box textAlign="center">
          <IconButton
            title={t('packages|Add File')}
            disabled={isEditing()}
            aria-label={t('packages|Add File')}
            onClick={() => addFile()}
          >
            <Icon icon={addIcon} width="15" height="15" />
          </IconButton>
        </Box>
      )}
    </List>
  );
}
