import { network, profile } from '../../wailsjs/go/models';
import type {
  StaticFormErrors,
  StaticFormField,
  StaticFormValues,
} from '../lib/validation';
import AdapterModeBadge from './AdapterModeBadge';
import StaticIpForm from './StaticIpForm';

interface IpConfigTabProps {
  adapters: network.Adapter[];
  selectedAdapter: string;
  onSelectAdapter: (name: string) => void;
  currentAdapter?: network.Adapter;
  onManageAdapters: () => void;

  isDHCP: boolean;
  onChangeMode: (isDHCP: boolean) => void;

  values: StaticFormValues;
  errors: StaticFormErrors;
  onChangeField: (field: StaticFormField, value: string) => void;

  profiles: profile.IPProfile[];
  selectedProfileId: string;
  onSelectProfile: (id: string) => void;
  onOpenProfileFolder: () => void;
  profileName: string;
  onChangeProfileName: (value: string) => void;
  onSaveProfile: () => void;
  onUpdateProfile: () => void;
  onDeleteProfile: () => void;
  isProfileDirty: boolean;

  isLoading: boolean;
  canApply: boolean;
  onApply: () => void;
}

export default function IpConfigTab({
  adapters,
  selectedAdapter,
  onSelectAdapter,
  currentAdapter,
  onManageAdapters,
  isDHCP,
  onChangeMode,
  values,
  errors,
  onChangeField,
  profiles,
  selectedProfileId,
  onSelectProfile,
  onOpenProfileFolder,
  profileName,
  onChangeProfileName,
  onSaveProfile,
  onUpdateProfile,
  onDeleteProfile,
  isProfileDirty,
  isLoading,
  canApply,
  onApply,
}: IpConfigTabProps) {
  return (
    <div className="flex flex-col">
      {/* Adapter Selection */}
      <div className="mb-4">
        <div className="flex justify-between items-center mb-1.5">
          <label
            htmlFor="adapter-select"
            className="block text-xs font-semibold text-gray-600 dark:text-gray-300"
          >
            Ağ Adaptörü
          </label>
          <button
            onClick={onManageAdapters}
            className="text-xs text-blue-600 dark:text-blue-400 hover:underline font-medium"
          >
            Tüm Adaptörleri Yönet
          </button>
        </div>
        <select
          id="adapter-select"
          className="w-full border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 p-2 border text-sm"
          value={selectedAdapter}
          onChange={(e) => onSelectAdapter(e.target.value)}
        >
          {adapters
            .filter((a) => a.enabled)
            .map((a, idx) => (
              <option key={idx} value={a.name}>
                {a.name}
              </option>
            ))}
        </select>
        {currentAdapter && (
          <div className="flex justify-between items-center mt-1.5">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              Mevcut IP:{' '}
              <span className="font-semibold text-gray-800 dark:text-gray-200 select-text">
                {currentAdapter.ipAddress || 'Bilinmiyor'}
              </span>
            </p>
            <AdapterModeBadge mode={currentAdapter.mode} />
          </div>
        )}
      </div>

      {/* Mode Selection */}
      <div className="flex bg-gray-100 dark:bg-gray-700/70 p-1 rounded-lg mb-4">
        <button
          className={`flex-1 py-1.5 rounded-md text-xs font-semibold transition ${
            !isDHCP
              ? 'bg-white dark:bg-gray-800 shadow text-blue-600 dark:text-blue-400'
              : 'text-gray-500 dark:text-gray-400'
          }`}
          aria-pressed={!isDHCP}
          onClick={() => onChangeMode(false)}
        >
          Statik IP
        </button>
        <button
          className={`flex-1 py-1.5 rounded-md text-xs font-semibold transition ${
            isDHCP
              ? 'bg-white dark:bg-gray-800 shadow text-blue-600 dark:text-blue-400'
              : 'text-gray-500 dark:text-gray-400'
          }`}
          aria-pressed={isDHCP}
          onClick={() => onChangeMode(true)}
        >
          DHCP (Otomatik)
        </button>
      </div>

      {/* Static IP Form */}
      {!isDHCP && (
        <StaticIpForm
          values={values}
          errors={errors}
          onChangeField={onChangeField}
          profiles={profiles}
          selectedProfileId={selectedProfileId}
          onSelectProfile={onSelectProfile}
          onOpenProfileFolder={onOpenProfileFolder}
          profileName={profileName}
          onChangeProfileName={onChangeProfileName}
          onSaveProfile={onSaveProfile}
          onUpdateProfile={onUpdateProfile}
          onDeleteProfile={onDeleteProfile}
          isProfileDirty={isProfileDirty}
        />
      )}

      {/* Apply Button */}
      <button
        onClick={onApply}
        disabled={isLoading || !canApply}
        title={
          !canApply && !isLoading
            ? 'IP, Subnet ve Gateway alanlarını geçerli değerlerle doldurun'
            : undefined
        }
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-2.5 px-4 rounded-lg shadow-md transition disabled:opacity-50 disabled:cursor-not-allowed mt-1 text-sm"
      >
        {isLoading ? 'İşleniyor...' : 'Ağa Uygula (Apply)'}
      </button>
    </div>
  );
}
