import { Component } from 'react';
import { ErrorState } from './UI';

export default class ErrorBoundary extends Component {
  state = { error: null };

  static getDerivedStateFromError(error) {
    return { error };
  }

  componentDidUpdate(previousProps) {
    if (previousProps.resetKey !== this.props.resetKey && this.state.error) {
      this.setState({ error: null });
    }
  }

  render() {
    if (this.state.error) {
      return <ErrorState error={this.state.error} retry={() => this.setState({ error: null })} />;
    }
    return this.props.children;
  }
}
