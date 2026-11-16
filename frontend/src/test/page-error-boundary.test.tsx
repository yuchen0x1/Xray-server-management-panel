import { fireEvent, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it, vi } from 'vitest';

import { PageErrorBoundaryBase } from '@/components/feedback/PageErrorBoundary';
import { renderWithProviders } from './test-utils';

function BrokenPage(): never {
  throw new Error('render failure');
}

describe('PageErrorBoundary', () => {
  it('keeps the panel shell and offers recovery actions after a page render error', () => {
    const onHome = vi.fn();
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    renderWithProviders(
      <MemoryRouter>
        <PageErrorBoundaryBase
          homeLabel="Dashboard"
          onHome={onHome}
          retryLabel="Retry"
          title="Page failed"
        >
          <BrokenPage />
        </PageErrorBoundaryBase>
      </MemoryRouter>,
    );

    expect(screen.getByText('Page failed')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Dashboard' }));
    expect(onHome).toHaveBeenCalledOnce();

    consoleError.mockRestore();
  });
});
