import addIcon from '@iconify/icons-mdi/plus';
import Icon from '@iconify/react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import List from '@mui/material/List';
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { File, Package } from '../../api/apiDataTypes';
import ListItem from '../common/ListItem';
import MoreMenu from '../common/MoreMenu';

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
                {t('frequent|size', { size: hasSize() ? fileToEdit.size : '-' })}
              </Typography>
            </Grid>
            <Grid item>
              <Typography component="span" variant="body2" color="textSecondary">
                {`${t('packages|sha1_hash_base64')}: ${fileToEdit.hash || '-'}`}
              </Typography>
            </Grid>
            <Grid item>
              <Typography component="span" variant="body2" color="textSecondary">
                {`${t('packages|sha256_hash_hex')}: ${fileToEdit.hash256 || '-'}`}
              </Typography>
            </Grid>
          </Grid>
        </Grid>
        <ListItemSecondaryAction>
          <MoreMenu
            iconButtonProps={{ disabled: showEditOptions }}
            options={[
              {
                label: t('frequent|edit'),
                action: () => {
                  onEditClicked?.();
                },
              },
              {
                label: t('frequent|delete'),
                action: () => {
                  onDeleteClicked?.();
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
            {!file.name ? t('packages|new_file') : t('packages|edit_file')}
          </Typography>
        </Grid>
        <Grid item xs={12}>
          <TextField
            name="name"
            margin="dense"
            label={t('packages|name')}
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
              label={t('packages|size')}
              type="text"
              helperText={t('packages|in_bytes')}
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
              label={t('packages|sha1_hash_base64')}
              type="text"
              helperText={t('packages|tip_command', {
                command: 'cat FILE | openssl dgst -sha1 -binary | base64',
              })}
              fullWidth
              value={fileToEdit.hash}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
                onFileInfoChanged('hash', e.target.value)
              }
            />
          </Grid>
          <Grid item xs={6}>
            <TextField
              name="hash256"
              margin="dense"
              label={t('packages|sha256_hash_hex')}
              type="text"
              helperText={t('packages|tip_command', {
                command: 'sha256sum FILE',
              })}
              fullWidth
              value={fileToEdit.hash256}
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
                onFileInfoChanged('hash256', e.target.value)
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
              {t('frequent|cancel')}
            </Button>
          </Grid>
          <Grid item>
            <Button onClick={() => onEditFinished(fileToEdit)}>{t('frequent|done')}</Button>
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
          onEditFinished={(file?: File) => (file ? onFileChanged(file, i) : stopEditing())}
          edit={editFileIndex === i}
          onEditClicked={() => onEditClicked(i)}
          onDeleteClicked={() => deleteFile(i)}
        />
      ))}
      {showFileAdd ? (
        <FileListItem
          file={{ name: '', size: '', hash: '', hash256: '' }}
          onEditFinished={(file?: File) => addFile(file)}
          edit
        />
      ) : (
        <Box textAlign="center">
          <IconButton
            title={t('packages|add_file')}
            disabled={isEditing()}
            aria-label={t('packages|add_file')}
            onClick={() => addFile()}
            size="large"
          >
            <Icon icon={addIcon} width="15" height="15" />
          </IconButton>
        </Box>
      )}
    </List>
  );
}
