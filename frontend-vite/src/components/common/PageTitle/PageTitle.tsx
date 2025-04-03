import React from 'react';

export default function PageTitle({
  title,
  children,
}: {
  title: string | null | undefined;
  children?: React.ReactNode;
}) {
  React.useEffect(() => {
    document.title = title || '';
  }, [title]);

  return <>{children}</>;
}
