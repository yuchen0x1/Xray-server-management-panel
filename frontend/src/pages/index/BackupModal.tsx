import { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Alert, App, Button, Modal } from 'antd';
import { DownloadOutlined, UploadOutlined } from '@ant-design/icons';

import { HttpUtil, PromiseUtil } from '@/utils';
import { withBasePath } from '@/utils/base-path';
import './BackupModal.css';

interface BusyEvent {
  busy: boolean;
  tip?: string;
}

interface BackupModalProps {
  open: boolean;
  basePath: string;
  onClose: () => void;
  onBusy: (e: BusyEvent) => void;
}

export default function BackupModal({ open, basePath: _basePath, onClose, onBusy }: BackupModalProps) {
  const { t } = useTranslation();
  const { message } = App.useApp();
  const fileInput = useRef<HTMLInputElement>(null);
  const restoring = useRef(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const isPostgres = window.X_UI_DB_TYPE === 'postgres';

  function exportDb() {
    window.location.href = withBasePath('/panel/api/server/getDb');
  }

  function exportMigration() {
    window.location.href = withBasePath('/panel/api/server/getMigration');
  }

  async function importDb() {
    if (!selectedFile || restoring.current) return;
    restoring.current = true;
    setBusy(true);
    try {
      const formData = new FormData();
      formData.append('db', selectedFile);

      onBusy({ busy: true, tip: `${t('pages.index.importDatabase')}…` });

      const upload = await HttpUtil.post('/panel/api/server/importDB', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      if (!upload?.success) return;

      setSelectedFile(null);
      onClose();

      onBusy({ busy: true, tip: `${t('pages.settings.restartPanel')}…` });
      const restart = await HttpUtil.post('/panel/api/setting/restartPanel');
      if (restart?.success) {
        await PromiseUtil.sleep(5000);
        window.location.reload();
      }
    } catch {
      void message.error(t('pages.index.importDatabaseError'));
    } finally {
      restoring.current = false;
      setBusy(false);
      onBusy({ busy: false });
    }
  }

  return (
    <Modal
      open={open}
      title={t(selectedFile ? 'pages.index.restoreConfirmTitle' : 'pages.index.backupTitle')}
      footer={selectedFile ? undefined : null}
      okText={t('pages.index.importDatabase')}
      cancelText={t('cancel')}
      okButtonProps={{ danger: true }}
      confirmLoading={busy}
      cancelButtonProps={{ disabled: busy }}
      onOk={() => { void importDb(); }}
      onCancel={() => {
        if (restoring.current) return;
        if (selectedFile) setSelectedFile(null);
        else onClose();
      }}
      closable={!busy}
      keyboard={!busy}
      mask={{ closable: !busy }}
    >
      <input
        ref={fileInput}
        type="file"
        accept=".dump,.db"
        hidden
        aria-label={t('pages.index.importDatabase')}
        onChange={(event) => {
          const file = event.target.files?.[0];
          event.target.value = '';
          if (!file || restoring.current) return;
          if (file.size === 0) {
            void message.error(t('pages.index.restoreEmptyFile'));
            return;
          }
          setSelectedFile(file);
        }}
      />
      {selectedFile ? (
        <>
          <Alert type="warning" showIcon title={t('pages.index.restoreConfirmWarning')} />
          <p className="backup-file-name">{selectedFile.name}</p>
        </>
      ) : (
        <>
          {isPostgres && (
            <div className="backup-description" style={{ marginBottom: 16 }}>
              {t('pages.index.backupPostgresNote')}
            </div>
          )}
          <div className="backup-list">
            <div className="backup-item">
              <div className="backup-meta">
                <div className="backup-title">{t('pages.index.exportDatabase')}</div>
                <div className="backup-description">
                  {isPostgres ? t('pages.index.exportDatabasePgDesc') : t('pages.index.exportDatabaseDesc')}
                </div>
              </div>
              <Button type="primary" aria-label={t('pages.index.exportDatabase')} onClick={exportDb} icon={<DownloadOutlined />} />
            </div>

            {isPostgres && (
              <div className="backup-item">
                <div className="backup-meta">
                  <div className="backup-title">{t('pages.index.migrationDownload')}</div>
                  <div className="backup-description">{t('pages.index.migrationDownloadPgDesc')}</div>
                </div>
                <Button type="primary" aria-label={t('pages.index.migrationDownload')} onClick={exportMigration} icon={<DownloadOutlined />} />
              </div>
            )}

            <div className="backup-item">
              <div className="backup-meta">
                <div className="backup-title">{t('pages.index.importDatabase')}</div>
                <div className="backup-description">
                  {isPostgres ? t('pages.index.importDatabasePgDesc') : t('pages.index.importDatabaseDesc')}
                </div>
              </div>
              <Button danger type="primary" disabled={busy} aria-label={t('pages.index.importDatabase')} onClick={() => fileInput.current?.click()} icon={<UploadOutlined />} />
            </div>
          </div>
        </>
      )}
    </Modal>
  );
}
