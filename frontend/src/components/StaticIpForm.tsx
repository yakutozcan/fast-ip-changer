import { profile } from '../../wailsjs/go/models';
import type {
  StaticFormErrors,
  StaticFormField,
  StaticFormValues,
} from '../lib/validation';

interface StaticIpFormProps {
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
}

const FIELD_CLASSES =
  'w-full rounded-md shadow-sm p-2 border text-xs font-mono dark:bg-gray-700 dark:text-white select-text';

function fieldClassName(hasError: boolean): string {
  return `${FIELD_CLASSES} ${
    hasError
      ? 'border-red-500 dark:border-red-500 focus:border-red-500 focus:ring-red-500'
      : 'border-gray-300 dark:border-gray-600'
  }`;
}

interface IpFieldProps {
  id: string;
  label: string;
  placeholder: string;
  value: string;
  error?: string;
  onChange: (value: string) => void;
}

function IpField({ id, label, placeholder, value, error, onChange }: IpFieldProps) {
  return (
    <div>
      <label htmlFor={id} className="block text-[11px] text-gray-500 dark:text-gray-400 mb-1">
        {label}
      </label>
      <input
        id={id}
        type="text"
        inputMode="decimal"
        autoComplete="off"
        spellCheck={false}
        className={fieldClassName(Boolean(error))}
        value={value}
        placeholder={placeholder}
        aria-invalid={Boolean(error)}
        aria-describedby={error ? `${id}-error` : undefined}
        onChange={(e) => onChange(e.target.value)}
      />
      {error && (
        <p id={`${id}-error`} className="mt-1 text-[10px] font-medium text-red-600 dark:text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}

export default function StaticIpForm({
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
}: StaticIpFormProps) {
  const hasSelectedProfile = Boolean(selectedProfileId);

  return (
    <div className="space-y-3 mb-4">
      {/* Saved profiles */}
      <div>
        <div className="flex justify-between items-center mb-1.5">
          <label
            htmlFor="profile-select"
            className="block text-xs font-semibold text-gray-600 dark:text-gray-300"
          >
            Kayıtlı Profiller
          </label>
          <button
            onClick={onOpenProfileFolder}
            className="text-xs text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 transition flex items-center"
            title="profiles.json dosyasının bulunduğu klasörü aç"
          >
            <span aria-hidden="true" className="mr-1">
              📁
            </span>{' '}
            Klasörü Aç
          </button>
        </div>
        <div className="flex space-x-2">
          <select
            id="profile-select"
            className="flex-1 border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm p-2 border text-sm"
            value={selectedProfileId}
            onChange={(e) => onSelectProfile(e.target.value)}
          >
            <option value="">-- Yeni IP Ayarı Gir --</option>
            {profiles.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          {hasSelectedProfile && (
            <button
              onClick={onDeleteProfile}
              className="px-3 bg-red-100 text-red-600 dark:bg-red-900/40 dark:text-red-300 rounded-md hover:bg-red-200 dark:hover:bg-red-800/60 transition text-xs font-medium"
            >
              Sil
            </button>
          )}
        </div>
      </div>

      {/* Address fields */}
      <div className="grid grid-cols-2 gap-3">
        <IpField
          id="field-ip"
          label="IP Adresi"
          placeholder="192.168.1.100"
          value={values.ip}
          error={errors.ip}
          onChange={(v) => onChangeField('ip', v)}
        />
        <IpField
          id="field-subnet"
          label="Subnet Mask"
          placeholder="255.255.255.0"
          value={values.subnet}
          error={errors.subnet}
          onChange={(v) => onChangeField('subnet', v)}
        />
        <IpField
          id="field-gateway"
          label="Default Gateway"
          placeholder="192.168.1.1"
          value={values.gateway}
          error={errors.gateway}
          onChange={(v) => onChangeField('gateway', v)}
        />
        <IpField
          id="field-dns"
          label="DNS (Opsiyonel)"
          placeholder="1.1.1.1, 8.8.8.8"
          value={values.dns}
          error={errors.dns}
          onChange={(v) => onChangeField('dns', v)}
        />
      </div>

      <p className="text-[10px] leading-snug text-gray-500 dark:text-gray-400">
        DNS alanı doldurulursa ağa uygulanır; birden fazla sunucuyu virgülle ayırın. Boş
        bırakılırsa mevcut DNS ayarı korunur.
      </p>

      {/* Profile save / update */}
      <div className="flex space-x-2 pt-3 border-t border-gray-200 dark:border-gray-700 mt-2">
        <label htmlFor="profile-name" className="sr-only">
          Profil Adı
        </label>
        <input
          id="profile-name"
          type="text"
          placeholder="Profil Adı (Örn: Ofis 192)"
          className="flex-1 border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm p-2 border text-xs select-text"
          value={profileName}
          onChange={(e) => onChangeProfileName(e.target.value)}
        />
        {hasSelectedProfile ? (
          <button
            onClick={onUpdateProfile}
            disabled={!isProfileDirty}
            title={isProfileDirty ? 'Profili güncelle' : 'Değişiklik yok'}
            className="px-3.5 bg-amber-600 text-white rounded-md hover:bg-amber-700 text-xs font-semibold transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Güncelle
          </button>
        ) : (
          <button
            onClick={onSaveProfile}
            className="px-3.5 bg-green-600 text-white rounded-md hover:bg-green-700 text-xs font-semibold transition"
          >
            Profili Kaydet
          </button>
        )}
      </div>
    </div>
  );
}
