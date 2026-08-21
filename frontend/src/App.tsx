import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  DisableAdapter,
  EnableAdapter,
  OpenProfileFolder,
  QuickCheck,
  SetDHCP,
  SetStaticIP,
} from '../wailsjs/go/main/App';
import { diagnostics, network, profile } from '../wailsjs/go/models';

import AdaptersModal from './components/AdaptersModal';
import AppHeader from './components/AppHeader';
import ConnectionStatusBar from './components/ConnectionStatusBar';
import DiagnosticsTab from './components/DiagnosticsTab';
import ElevationBanner from './components/ElevationBanner';
import IpConfigTab from './components/IpConfigTab';
import MessageBanner from './components/MessageBanner';
import TabNav, { type TabId } from './components/TabNav';

import { useAdapters } from './hooks/useAdapters';
import { useContextMenuGuard } from './hooks/useContextMenuGuard';
import { useDarkMode } from './hooks/useDarkMode';
import { useDiagnostics } from './hooks/useDiagnostics';
import { useElevation } from './hooks/useElevation';
import { useMessage } from './hooks/useMessage';
import { useProfiles } from './hooks/useProfiles';

import { STORAGE_KEYS, readStored, writeStored } from './lib/storage';
import {
  isStaticFormValid,
  normalizeDnsList,
  validateStaticForm,
  type StaticFormField,
  type StaticFormValues,
} from './lib/validation';

const EMPTY_FORM: StaticFormValues = {
  ip: '',
  subnet: '255.255.255.0',
  gateway: '',
  dns: '',
};

export default function App() {
  const { darkMode, toggleDarkMode } = useDarkMode();
  const { message, showMessage, setPersistentMessage } = useMessage();
  useContextMenuGuard();

  const isElevated = useElevation();
  const [activeTab, setActiveTab] = useState<TabId>('config');

  const {
    adapters,
    selectedAdapter,
    setSelectedAdapter,
    currentAdapter,
    isRefreshing: isRefreshingAdapters,
    reload: reloadAdapters,
  } = useAdapters(showMessage);

  const { profiles, reload: reloadProfiles, createProfile, updateProfile, deleteProfile } =
    useProfiles();

  const [isDHCP, setIsDHCP] = useState(false);
  const [form, setForm] = useState<StaticFormValues>(EMPTY_FORM);
  const [selectedProfileId, setSelectedProfileId] = useState('');
  const [profileName, setProfileName] = useState('');

  const diag = useDiagnostics(showMessage);

  const [quickStatus, setQuickStatus] = useState<diagnostics.QuickCheckResult | null>(null);
  const [isCheckingStatus, setIsCheckingStatus] = useState(false);
  const [checkPublicIP, setCheckPublicIP] = useState<boolean>(
    () => readStored(STORAGE_KEYS.publicIpLookup) === 'true',
  );

  const [isLoading, setIsLoading] = useState(false);
  const [showAllAdapters, setShowAllAdapters] = useState(false);

  // Latest values for QuickCheck defaults, so runQuickCheck stays stable.
  const quickCheckDefaults = useRef({ gateway: EMPTY_FORM.gateway, checkPublicIP: false });
  useEffect(() => {
    quickCheckDefaults.current = { gateway: form.gateway, checkPublicIP };
  }, [form.gateway, checkPublicIP]);

  const runQuickCheck = useCallback(async (gw?: string, withPublicIp?: boolean) => {
    setIsCheckingStatus(true);
    try {
      const defaults = quickCheckDefaults.current;
      const res = await QuickCheck(
        gw !== undefined ? gw : defaults.gateway,
        withPublicIp !== undefined ? withPublicIp : defaults.checkPublicIP,
      );
      setQuickStatus(res);
    } catch (err) {
      console.error(err);
    } finally {
      setIsCheckingStatus(false);
    }
  }, []);

  // Initialize
  useEffect(() => {
    void reloadAdapters();
    void reloadProfiles();
    void runQuickCheck();
  }, [reloadAdapters, reloadProfiles, runQuickCheck]);

  const handleToggleCheckPublicIP = useCallback(
    (next: boolean) => {
      setCheckPublicIP(next);
      writeStored(STORAGE_KEYS.publicIpLookup, String(next));
      if (!next) {
        setQuickStatus((prev) =>
          prev ? diagnostics.QuickCheckResult.createFrom({ ...prev, publicIp: '' }) : prev,
        );
      }
      void runQuickCheck(undefined, next);
    },
    [runQuickCheck],
  );

  const handleChangeField = useCallback((field: StaticFormField, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  }, []);

  /** Single reset path so every field (subnet included) is cleared uniformly. */
  const clearProfileSelection = useCallback((resetForm: boolean) => {
    setSelectedProfileId('');
    setProfileName('');
    if (resetForm) setForm(EMPTY_FORM);
  }, []);

  const handleProfileSelect = useCallback(
    (id: string) => {
      if (!id) {
        clearProfileSelection(true);
        return;
      }
      const p = profiles.find((x) => x.id === id);
      setSelectedProfileId(id);
      if (p) {
        setProfileName(p.name);
        setForm({ ip: p.ip, subnet: p.subnet, gateway: p.gateway, dns: p.dns });
      }
    },
    [clearProfileSelection, profiles],
  );

  const selectedProfile = useMemo(
    () => profiles.find((p) => p.id === selectedProfileId),
    [profiles, selectedProfileId],
  );

  const isProfileDirty = useMemo(() => {
    if (!selectedProfile) return false;
    return (
      profileName.trim() !== selectedProfile.name ||
      form.ip !== selectedProfile.ip ||
      form.subnet !== selectedProfile.subnet ||
      form.gateway !== selectedProfile.gateway ||
      form.dns !== selectedProfile.dns
    );
  }, [form, profileName, selectedProfile]);

  const formErrors = useMemo(() => validateStaticForm(form), [form]);
  const canApply = isDHCP || isStaticFormValid(form);

  const handleSaveProfile = useCallback(async () => {
    if (!profileName.trim() || !form.ip || !form.subnet || !form.gateway) {
      showMessage('Lütfen profil adını ve gerekli IP alanlarını doldurun', 'error');
      return;
    }

    const newProfile: profile.IPProfile = {
      id: Date.now().toString(),
      name: profileName.trim(),
      ip: form.ip,
      subnet: form.subnet,
      gateway: form.gateway,
      dns: form.dns,
    };

    try {
      await createProfile(newProfile);
      setSelectedProfileId(newProfile.id);
      setProfileName(newProfile.name);
      showMessage('Profil başarıyla kaydedildi!', 'success');
    } catch (err) {
      showMessage('Profil kaydedilemedi: ' + String(err), 'error');
    }
  }, [createProfile, form, profileName, showMessage]);

  const handleUpdateProfile = useCallback(async () => {
    if (!selectedProfile) return;
    if (!profileName.trim() || !form.ip || !form.subnet || !form.gateway) {
      showMessage('Lütfen profil adını ve gerekli IP alanlarını doldurun', 'error');
      return;
    }

    try {
      await updateProfile({
        id: selectedProfile.id,
        name: profileName.trim(),
        ip: form.ip,
        subnet: form.subnet,
        gateway: form.gateway,
        dns: form.dns,
      });
      showMessage('Profil güncellendi', 'success');
    } catch (err) {
      showMessage('Profil güncellenemedi: ' + String(err), 'error');
    }
  }, [form, profileName, selectedProfile, showMessage, updateProfile]);

  const handleDeleteProfile = useCallback(async () => {
    if (!selectedProfileId) return;
    try {
      await deleteProfile(selectedProfileId);
      clearProfileSelection(true);
      showMessage('Profil silindi', 'success');
    } catch (err) {
      showMessage('Profil silinemedi: ' + String(err), 'error');
    }
  }, [clearProfileSelection, deleteProfile, selectedProfileId, showMessage]);

  const handleOpenProfileFolder = useCallback(async () => {
    try {
      await OpenProfileFolder();
    } catch (err) {
      console.error(err);
      showMessage('Klasör açılamadı', 'error');
    }
  }, [showMessage]);

  const handleApply = useCallback(async () => {
    if (!selectedAdapter) {
      showMessage('Lütfen bir ağ adaptörü seçin', 'error');
      return;
    }

    setIsLoading(true);
    setPersistentMessage('Uygulanıyor...', 'info');

    try {
      if (isDHCP) {
        await SetDHCP(selectedAdapter);
        showMessage('DHCP ayarları başarıyla uygulandı.', 'success');
      } else {
        if (!form.ip || !form.subnet || !form.gateway) {
          showMessage('IP, Subnet ve Gateway alanları zorunludur.', 'error');
          setIsLoading(false);
          return;
        }
        await SetStaticIP(
          selectedAdapter,
          form.ip.trim(),
          form.subnet.trim(),
          form.gateway.trim(),
          normalizeDnsList(form.dns),
        );
        showMessage('Statik IP ayarları başarıyla uygulandı.', 'success');
      }
      await reloadAdapters();
      void runQuickCheck(form.gateway.trim());
    } catch (err) {
      showMessage('Hata: ' + String(err), 'error');
    } finally {
      setIsLoading(false);
    }
  }, [
    form,
    isDHCP,
    reloadAdapters,
    runQuickCheck,
    selectedAdapter,
    setPersistentMessage,
    showMessage,
  ]);

  const handleToggleAdapterStatus = useCallback(
    async (adapter: network.Adapter) => {
      setIsLoading(true);
      setPersistentMessage('İşleniyor...', 'info');
      try {
        if (adapter.enabled) {
          await DisableAdapter(adapter.name);
          showMessage(adapter.name + ' devre dışı bırakıldı.', 'success');
        } else {
          await EnableAdapter(adapter.name);
          showMessage(adapter.name + ' etkinleştirildi.', 'success');
        }
        await reloadAdapters();
      } catch (err) {
        showMessage('İşlem başarısız: ' + String(err), 'error');
      } finally {
        setIsLoading(false);
      }
    },
    [reloadAdapters, setPersistentMessage, showMessage],
  );

  const closeAdaptersModal = useCallback(() => setShowAllAdapters(false), []);
  const refreshAdapters = useCallback(() => void reloadAdapters(), [reloadAdapters]);

  return (
    <div className="min-h-screen p-4 font-sans flex flex-col items-center select-none cursor-default bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100 relative">
      <div className="w-full max-w-md bg-white dark:bg-gray-800 rounded-xl shadow-lg p-5 transition-colors flex flex-col">
        <AppHeader darkMode={darkMode} onToggleDarkMode={toggleDarkMode} />

        {isElevated === false && <ElevationBanner />}

        <ConnectionStatusBar
          quickStatus={quickStatus}
          isCheckingStatus={isCheckingStatus}
          onRefresh={() => void runQuickCheck()}
          checkPublicIP={checkPublicIP}
          onToggleCheckPublicIP={handleToggleCheckPublicIP}
        />

        <TabNav activeTab={activeTab} onChangeTab={setActiveTab} />

        {activeTab === 'config' && (
          <div id="tabpanel-config" role="tabpanel" aria-labelledby="tab-config">
            <IpConfigTab
              adapters={adapters}
              selectedAdapter={selectedAdapter}
              onSelectAdapter={setSelectedAdapter}
              currentAdapter={currentAdapter}
              onManageAdapters={() => setShowAllAdapters(true)}
              isDHCP={isDHCP}
              onChangeMode={setIsDHCP}
              values={form}
              errors={formErrors}
              onChangeField={handleChangeField}
              profiles={profiles}
              selectedProfileId={selectedProfileId}
              onSelectProfile={handleProfileSelect}
              onOpenProfileFolder={handleOpenProfileFolder}
              profileName={profileName}
              onChangeProfileName={setProfileName}
              onSaveProfile={handleSaveProfile}
              onUpdateProfile={handleUpdateProfile}
              onDeleteProfile={handleDeleteProfile}
              isProfileDirty={isProfileDirty}
              isLoading={isLoading}
              canApply={canApply}
              onApply={handleApply}
            />
          </div>
        )}

        {activeTab === 'tools' && (
          <div id="tabpanel-tools" role="tabpanel" aria-labelledby="tab-tools">
            <DiagnosticsTab
              targetHost={diag.targetHost}
              onChangeTargetHost={diag.setTargetHost}
              pingCount={diag.pingCount}
              onChangePingCount={diag.setPingCount}
              gateway={form.gateway}
              consoleOutput={diag.consoleOutput}
              isExecutingTool={diag.isExecutingTool}
              isCancelling={diag.isCancelling}
              onPing={() => void diag.runPing()}
              onTraceRoute={() => void diag.runTraceRoute()}
              onCancel={() => void diag.cancel()}
              onCopyOutput={() => void diag.copyOutput()}
              onClearOutput={diag.clearOutput}
            />
          </div>
        )}

        <MessageBanner message={message} />
      </div>

      {showAllAdapters && (
        <AdaptersModal
          adapters={adapters}
          isLoading={isLoading}
          isRefreshing={isRefreshingAdapters}
          onToggleAdapter={handleToggleAdapterStatus}
          onRefresh={refreshAdapters}
          onClose={closeAdaptersModal}
        />
      )}
    </div>
  );
}
