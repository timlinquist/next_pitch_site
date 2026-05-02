import React from 'react';

const LocationMapEmbed = ({ address }) => {
  if (!address) return null;
  const encoded = encodeURIComponent(address);
  const embedSrc = `https://www.google.com/maps?q=${encoded}&output=embed`;
  const viewHref = `https://www.google.com/maps/search/?api=1&query=${encoded}`;

  return (
    <div className="location-map-embed">
      <iframe
        title={`Map of ${address}`}
        src={embedSrc}
        width="100%"
        height="300"
        style={{ border: 0 }}
        loading="lazy"
        referrerPolicy="no-referrer-when-downgrade"
      />
      <a
        className="location-map-link"
        href={viewHref}
        target="_blank"
        rel="noopener noreferrer"
      >
        Open in Google Maps ↗
      </a>
    </div>
  );
};

export default LocationMapEmbed;
