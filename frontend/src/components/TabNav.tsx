export type TabId = 'config' | 'tools';

interface TabNavProps {
  activeTab: TabId;
  onChangeTab: (tab: TabId) => void;
}

const TABS: { id: TabId; icon: string; label: string }[] = [
  { id: 'config', icon: '⚙️', label: 'IP Yapılandırma' },
  { id: 'tools', icon: '📡', label: 'Ping & Traceroute' },
];

export default function TabNav({ activeTab, onChangeTab }: TabNavProps) {
  return (
    <div role="tablist" className="flex border-b border-gray-200 dark:border-gray-700 mb-4">
      {TABS.map((tab) => (
        <button
          key={tab.id}
          role="tab"
          id={`tab-${tab.id}`}
          aria-selected={activeTab === tab.id}
          aria-controls={`tabpanel-${tab.id}`}
          onClick={() => onChangeTab(tab.id)}
          className={`flex-1 py-2 text-center text-xs font-semibold border-b-2 transition ${
            activeTab === tab.id
              ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400'
              : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
          }`}
        >
          <span aria-hidden="true">{tab.icon}</span> {tab.label}
        </button>
      ))}
    </div>
  );
}
