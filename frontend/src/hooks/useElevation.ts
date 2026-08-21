import { useEffect, useState } from 'react';
import { IsElevated } from '../../wailsjs/go/main/App';

/**
 * Whether the process has administrator / root rights.
 * `null` while unknown (still loading, or the check itself failed).
 */
export function useElevation(): boolean | null {
  const [isElevated, setIsElevated] = useState<boolean | null>(null);

  useEffect(() => {
    let cancelled = false;
    IsElevated()
      .then((value) => {
        if (!cancelled) setIsElevated(value);
      })
      .catch((err) => {
        console.error(err);
        if (!cancelled) setIsElevated(null);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return isElevated;
}
