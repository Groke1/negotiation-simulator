import { createContext, useContext, useEffect, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { getMe, switchRole } from '../api';
import { tokenStore } from '../api/client';
import { normalizeRoles, resolveActiveRole } from './roles';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const client = useQueryClient();
  const [signedOut, setSignedOut] = useState(!tokenStore.get());
  const meQuery = useQuery({ queryKey: ['me'], queryFn: ({ signal }) => getMe(signal), enabled: !signedOut, retry: false, staleTime: 5 * 60 * 1000 });

  useEffect(() => { if (meQuery.data?.id) localStorage.setItem('user_id', meQuery.data.id); }, [meQuery.data]);
  const normalizedMe = meQuery.data ? { ...meQuery.data, roles: normalizeRoles(meQuery.data.roles || []), active_role: resolveActiveRole(meQuery.data) } : undefined;

  useEffect(() => {
    const onUnauthorized = () => { setSignedOut(true); client.clear(); };
    window.addEventListener('app:unauthorized', onUnauthorized);
    return () => window.removeEventListener('app:unauthorized', onUnauthorized);
  }, [client]);

  const value = {
    me: normalizedMe,
    loading: !signedOut && meQuery.isLoading,
    error: meQuery.error,
    authenticated: !signedOut && Boolean(normalizedMe),
    async signInComplete() { setSignedOut(false); await client.invalidateQueries({ queryKey: ['me'] }); },
    signOut() { tokenStore.clear(); setSignedOut(true); client.clear(); },
    async setRole(role) { await switchRole(role); await client.invalidateQueries(); await client.refetchQueries({ queryKey: ['me'] }); },
  };
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => useContext(AuthContext);
