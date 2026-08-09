import { beforeEach, describe, expect, it } from 'vitest';

import { pauseAnimationsUntilLeave } from '@/hooks/useTheme';

describe('theme interaction helpers', () => {
  beforeEach(() => {
    document.documentElement.removeAttribute('data-theme-animations');
    document.body.innerHTML = '';
  });

  it('does not disable animations when the trigger is absent', () => {
    pauseAnimationsUntilLeave('missing-theme-trigger');

    expect(document.documentElement.hasAttribute('data-theme-animations')).toBe(false);
  });

  it('restores animations when the trigger loses focus', () => {
    const trigger = document.createElement('button');
    trigger.id = 'theme-trigger';
    document.body.appendChild(trigger);

    pauseAnimationsUntilLeave(trigger.id);
    expect(document.documentElement.getAttribute('data-theme-animations')).toBe('off');

    trigger.dispatchEvent(new FocusEvent('focusout'));

    expect(document.documentElement.hasAttribute('data-theme-animations')).toBe(false);
  });
});
