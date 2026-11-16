import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Button, Layout, Result, Space } from 'antd';
import { useLocation, useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';

import AppSidebar from '@/layouts/AppSidebar';

type PageErrorBoundaryProps = {
  children: ReactNode;
  homeLabel: string;
  retryLabel: string;
  title: string;
  onHome: () => void;
};

type PageErrorBoundaryState = {
  hasError: boolean;
};

export class PageErrorBoundaryBase extends Component<PageErrorBoundaryProps, PageErrorBoundaryState> {
  state: PageErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): PageErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(_error: Error, info: ErrorInfo) {
    console.error('Page render error', info.componentStack);
  }

  retry = () => {
    this.setState({ hasError: false });
  };

  render() {
    if (!this.state.hasError) return this.props.children;

    return (
      <Layout style={{ minHeight: '100vh', background: 'var(--glass-page-bg)' }}>
        <AppSidebar />
        <Layout className="content-shell" style={{ background: 'transparent' }}>
          <Layout.Content id="content-layout" className="content-area" style={{ padding: 28 }}>
            <Result
              status="error"
              title={this.props.title}
              extra={
                <Space wrap>
                  <Button type="primary" onClick={this.retry}>
                    {this.props.retryLabel}
                  </Button>
                  <Button onClick={this.props.onHome}>{this.props.homeLabel}</Button>
                </Space>
              }
            />
          </Layout.Content>
        </Layout>
      </Layout>
    );
  }
}

export default function PageErrorBoundary({ children }: { children: ReactNode }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <PageErrorBoundaryBase
      key={`${location.pathname}:${location.key}`}
      homeLabel={t('menu.dashboard')}
      retryLabel={t('refresh')}
      title={t('somethingWentWrong')}
      onHome={() => navigate('/')}
    >
      {children}
    </PageErrorBoundaryBase>
  );
}
