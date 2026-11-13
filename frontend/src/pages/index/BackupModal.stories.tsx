import { App } from 'antd';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import BackupModal from './BackupModal';

const meta = {
  title: 'Operations/BackupModal',
  component: BackupModal,
  decorators: [(Story) => <App><Story /></App>],
  args: {
    open: true,
    basePath: '/',
    onClose: fn(),
    onBusy: fn(),
  },
} satisfies Meta<typeof BackupModal>;

export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
