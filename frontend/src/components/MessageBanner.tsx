import type { AppMessage } from '../hooks/useMessage';

interface MessageBannerProps {
  message: AppMessage;
}

/**
 * The live region wrapper is always mounted so assistive tech announces
 * messages that appear later.
 */
export default function MessageBanner({ message }: MessageBannerProps) {
  return (
    <div role="status" aria-live="polite" aria-atomic="true">
      {message.text && (
        <div
          className={`p-2.5 mt-4 rounded-md text-xs font-medium select-text ${
            message.type === 'error'
              ? 'bg-red-100 text-red-700 dark:bg-red-900/60 dark:text-red-300'
              : message.type === 'success'
                ? 'bg-green-100 text-green-700 dark:bg-green-900/60 dark:text-green-300'
                : 'bg-blue-100 text-blue-700 dark:bg-blue-900/60 dark:text-blue-300'
          }`}
        >
          {message.text}
        </div>
      )}
    </div>
  );
}
