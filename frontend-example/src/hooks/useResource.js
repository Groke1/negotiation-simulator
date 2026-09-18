import { useCallback, useEffect, useState } from "react";

export function useResource(load, deps = []) {
  const [data, setData] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setData(await load());
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, deps);
  useEffect(() => {
    refresh();
  }, [refresh]);
  return { data, error, loading, refresh, setData };
}
