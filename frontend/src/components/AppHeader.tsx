interface AppHeaderProps {
  darkMode: boolean;
  onToggleDarkMode: () => void;
}

export default function AppHeader({ darkMode, onToggleDarkMode }: AppHeaderProps) {
  return (
    <div className="flex justify-between items-center mb-4">
      <div className="flex items-center space-x-2">
        <span aria-hidden="true" className="text-2xl">
          ⚡
        </span>
        <h1 className="text-xl font-bold tracking-tight text-gray-800 dark:text-white">
          Fast IP Changer
        </h1>
      </div>
      <button
        onClick={onToggleDarkMode}
        className="p-2 rounded-full bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition text-sm"
        title="Tema Değiştir"
        aria-label={darkMode ? 'Açık temaya geç' : 'Koyu temaya geç'}
        aria-pressed={darkMode}
      >
        <span aria-hidden="true">{darkMode ? '☀️' : '🌙'}</span>
      </button>
    </div>
  );
}
