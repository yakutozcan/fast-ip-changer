import { useCallback, useState } from 'react';
import {
  AddProfile,
  DeleteProfile,
  GetProfiles,
  UpdateProfile,
} from '../../wailsjs/go/main/App';
import { profile } from '../../wailsjs/go/models';

/** Saved IP profiles (profiles.json), with create / update / delete. */
export function useProfiles() {
  const [profiles, setProfiles] = useState<profile.IPProfile[]>([]);

  const reload = useCallback(async () => {
    try {
      const data = await GetProfiles();
      setProfiles(data || []);
    } catch (err) {
      console.error(err);
    }
  }, []);

  const createProfile = useCallback(
    async (p: profile.IPProfile) => {
      await AddProfile(p);
      await reload();
    },
    [reload],
  );

  const updateProfile = useCallback(
    async (p: profile.IPProfile) => {
      await UpdateProfile(p);
      await reload();
    },
    [reload],
  );

  const deleteProfile = useCallback(
    async (id: string) => {
      await DeleteProfile(id);
      await reload();
    },
    [reload],
  );

  return { profiles, reload, createProfile, updateProfile, deleteProfile };
}
