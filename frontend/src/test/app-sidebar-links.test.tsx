import { screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';

import AppSidebar from '@/layouts/AppSidebar';
import { renderWithProviders } from './test-utils';

describe('AppSidebar external actions', () => {
  it('exposes working documentation and donation links', () => {
    renderWithProviders(
      <MemoryRouter>
        <AppSidebar />
      </MemoryRouter>,
    );

    const docs = screen.getAllByRole('link', { name: 'Documentation' });
    const donations = screen.getAllByRole('link', { name: 'Donate' });

    expect(docs[0].getAttribute('href')).toBe('https://docs.sanaei.dev');
    expect(docs[0].getAttribute('target')).toBe('_blank');
    expect(docs[0].getAttribute('rel')).toBe('noreferrer');
    expect(donations[0].getAttribute('href')).toBe('https://donate.sanaei.dev/');
    expect(donations[0].getAttribute('target')).toBe('_blank');
    expect(donations[0].getAttribute('rel')).toBe('noreferrer');
  });
});
