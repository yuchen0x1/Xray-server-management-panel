import { App } from 'antd';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import BackupModal from '@/pages/index/BackupModal';
import { HttpUtil, Msg } from '@/utils';

const post = vi.mocked(HttpUtil.post);

function setup() {
  const onClose = vi.fn();
  const onBusy = vi.fn();
  const view = render(<App><BackupModal open basePath="/" onClose={onClose} onBusy={onBusy} /></App>);
  const input = document.querySelector<HTMLInputElement>('input[type="file"]');
  if (!input) throw new Error('Backup file input missing');
  const choose = (file = new File(['backup'], 'panel.db')) => {
    fireEvent.change(input, { target: { files: [file] } });
  };
  return { ...view, onClose, onBusy, choose };
}

async function confirmationDialog() {
  const title = await screen.findByText('Confirm database restore');
  const dialog = title.closest('[role="dialog"]');
  if (!(dialog instanceof HTMLElement)) throw new Error('Restore dialog missing');
  return dialog;
}

beforeEach(() => {
  post.mockReset();
});

describe('database restore safeguards', () => {
  it('requires confirmation and allows cancelling without uploading', async () => {
    const { choose, onBusy } = setup();
    choose();
    const dialog = await confirmationDialog();
    expect(within(dialog).getByText('panel.db')).toBeTruthy();
    expect(within(dialog).getByText(/replaces all current panel data/)).toBeTruthy();
    expect(post).not.toHaveBeenCalled();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));
    await waitFor(() => expect(screen.queryByText('Confirm database restore')).toBeNull(), { timeout: 5000 });
    expect(onBusy).not.toHaveBeenCalled();
    expect(post).not.toHaveBeenCalled();
  });

  it('rejects empty files before asking for confirmation', async () => {
    const { choose } = setup();
    choose(new File([], 'empty.db'));
    expect(await screen.findByText('The selected backup is empty. Choose a valid backup file.')).toBeTruthy();
    expect(screen.queryByText('Confirm database restore')).toBeNull();
    expect(post).not.toHaveBeenCalled();
  });

  it('prevents duplicate submissions and keeps failed uploads retryable', async () => {
    let finish: (msg: Msg) => void = () => {};
    post.mockImplementationOnce(() => new Promise<Msg>((resolve) => { finish = resolve; }));
    const { choose, onBusy, onClose } = setup();
    choose();
    const dialog = await confirmationDialog();
    const restore = within(dialog).getByRole('button', { name: 'Restore' });
    fireEvent.click(restore);
    fireEvent.click(restore);
    expect(post).toHaveBeenCalledTimes(1);
    expect(post.mock.calls[0][0]).toBe('/panel/api/server/importDB');
    const payload = post.mock.calls[0][1] as FormData;
    expect((payload.get('db') as File).name).toBe('panel.db');
    finish(new Msg(false, 'Invalid backup'));
    await waitFor(() => expect(onBusy).toHaveBeenLastCalledWith({ busy: false }));
    expect(onClose).not.toHaveBeenCalled();
    expect(screen.getByText('Confirm database restore')).toBeTruthy();
    post.mockResolvedValueOnce(new Msg(false, 'Retry rejected'));
    fireEvent.click(restore);
    await waitFor(() => expect(post).toHaveBeenCalledTimes(2));
  });

  it('clears the busy state after an unexpected failure', async () => {
    post.mockRejectedValueOnce(new Error('Disconnected'));
    const { choose, onBusy } = setup();
    choose();
    const dialog = await confirmationDialog();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Restore' }));
    await waitFor(() => expect(onBusy).toHaveBeenLastCalledWith({ busy: false }));
    expect(await screen.findByText('An error occurred while importing the database.')).toBeTruthy();
    expect(post).toHaveBeenCalledTimes(1);
  });

  it('restarts the panel only after a successful restore', async () => {
    post.mockResolvedValueOnce(new Msg(true)).mockResolvedValueOnce(new Msg(false, 'Restart failed'));
    const { choose, onBusy, onClose } = setup();
    choose();
    const dialog = await confirmationDialog();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Restore' }));
    await waitFor(() => expect(post).toHaveBeenNthCalledWith(2, '/panel/api/setting/restartPanel'));
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onBusy).toHaveBeenLastCalledWith({ busy: false });
  });
});
