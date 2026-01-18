export function useApiUrl() {
  return import.meta.env.VITE_API_URL || '/api';
}

export function useApi() {
  const apiUrl = useApiUrl();

  return {
    url: apiUrl,
    songs: `${apiUrl}/songs`,
    corrections: `${apiUrl}/corrections`,
    correctionsBatch: `${apiUrl}/corrections/batch`,
    adminPending: `${apiUrl}/admin/pending`,
    adminApprove: (id: number) => `${apiUrl}/admin/approve/${id}`,
    adminReject: (id: number) => `${apiUrl}/admin/reject/${id}`,
  };
}
