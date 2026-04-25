export function formatAddress(loc) {
  if (!loc) return '';
  const base = `${loc.street}, ${loc.city}, ${loc.state} ${loc.zip}`;
  if (loc.country && loc.country !== 'US') {
    return `${base}, ${loc.country}`;
  }
  return base;
}
