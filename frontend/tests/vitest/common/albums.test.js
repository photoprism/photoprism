import { describe, it, expect } from 'vitest';
import { processAlbumSelection } from '../../../src/common/albums.js';

function album(title, uid) {
  return { Title: title, UID: uid };
}

describe('processAlbumSelection', () => {
  it('trims whitespace and matches existing albums (case-insensitive)', () => {
    const available = [album('Summer', '1')];
    const selected = ['  summer  '];

    const { processed, changed } = processAlbumSelection(selected, available);

    expect(processed).toHaveLength(1);
    expect(processed[0]).toEqual(available[0]);
    expect(changed).toBe(true);
  });

  it('deduplicates identical UIDs and strings resolving to same album', () => {
    const a1 = album('Trips', 't1');
    const selected = [a1, a1, 'trips', 'TRIPS'];

    const { processed, changed } = processAlbumSelection(selected, [a1]);

    expect(processed).toHaveLength(1);
    expect(processed[0]).toEqual(a1);
    expect(changed).toBe(true);
  });

  it('keeps unmatched names as trimmed strings (for creation later)', () => {
    const selected = [' New Album  '];

    const { processed, changed } = processAlbumSelection(selected, []);

    expect(processed).toEqual(['New Album']);
    // No structural change: only trimming does not count as change if lengths are equal and no replacements/drops
    expect(changed).toBe(false);
  });

  it('drops empty / whitespace-only entries and reports change', () => {
    const selected = ['  ', '\n', '\t', '  Name  '];

    const { processed, changed } = processAlbumSelection(selected, []);

    expect(processed).toEqual(['Name']);
    expect(changed).toBe(true);
  });

  it('reconciles selection with new available items (race condition)', () => {
    const selected = ['  Road Trip '];
    const availableThen = [album('Road Trip', 'rt01')];

    const { processed, changed } = processAlbumSelection(selected, availableThen);

    expect(processed).toHaveLength(1);
    expect(processed[0]).toEqual(availableThen[0]);
    expect(changed).toBe(true);
  });

  it('preserves existing album objects and prevents duplicates', () => {
    const a = album('Family', 'fam1');
    const selected = [a, { ...a }]; // two distinct objects with same UID

    const { processed, changed } = processAlbumSelection(selected, [a]);

    expect(processed).toHaveLength(1);
    expect(processed[0]).toEqual(a);
    expect(changed).toBe(true);
  });
});
