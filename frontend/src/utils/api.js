// Get the base URL for API calls
export const getApiBaseUrl = () => {
    // In development, use the current hostname
    // In production, this will be the same as the frontend hostname
    return window.location.origin;
};

// Construct a full API URL
export const getApiUrl = (path) => {
    const baseUrl = getApiBaseUrl();
    // Remove any leading slash from the path to avoid double slashes
    const cleanPath = path.startsWith('/') ? path.slice(1) : path;
    return `${baseUrl}/api/${cleanPath}`;
}; 