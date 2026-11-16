import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const translationDir = path.resolve(scriptDir, '../../internal/web/translation');
const files = fs.readdirSync(translationDir).filter((file) => file.endsWith('.json')).sort();

const flatten = (value, prefix = '', result = {}) => {
  if (Array.isArray(value)) {
    value.forEach((item, index) => flatten(item, `${prefix}[${index}]`, result));
    return result;
  }
  if (value && typeof value === 'object') {
    for (const [key, child] of Object.entries(value)) {
      flatten(child, prefix ? `${prefix}.${key}` : key, result);
    }
    return result;
  }
  result[prefix] = value;
  return result;
};

const readLocale = (file) => {
  const fullPath = path.join(translationDir, file);
  return flatten(JSON.parse(fs.readFileSync(fullPath, 'utf8')));
};

if (!files.includes('en-US.json')) {
  throw new Error('en-US.json is required as the translation key source');
}

const source = readLocale('en-US.json');
const sourceKeys = Object.keys(source);
const sourceKeySet = new Set(sourceKeys);
const failures = [];
const placeholders = (value) => [...new Set((String(value).match(/\{[A-Za-z0-9_.-]+\}/g) ?? []))].sort();

for (const file of files.filter((entry) => entry !== 'en-US.json')) {
  const locale = readLocale(file);
  const missing = sourceKeys.filter((key) => !(key in locale) || locale[key] === '');
  const extra = Object.keys(locale).filter((key) => !sourceKeySet.has(key));
  const placeholderMismatches = sourceKeys
    .filter((key) => key in locale && JSON.stringify(placeholders(source[key])) !== JSON.stringify(placeholders(locale[key])))
    .map((key) => `${key} (${placeholders(source[key]).join(',')} -> ${placeholders(locale[key]).join(',')})`);
  if (missing.length > 0) failures.push(`${file} missing or empty: ${missing.join(', ')}`);
  if (extra.length > 0) failures.push(`${file} extra: ${extra.join(', ')}`);
  if (placeholderMismatches.length > 0) failures.push(`${file} placeholder mismatch: ${placeholderMismatches.join(', ')}`);
}

if (failures.length > 0) {
  console.error('Translation key check failed:');
  failures.forEach((failure) => console.error(`- ${failure}`));
  process.exitCode = 1;
} else {
  console.log(`Translation key check passed for ${files.length} locales and ${sourceKeys.length} keys.`);
}
