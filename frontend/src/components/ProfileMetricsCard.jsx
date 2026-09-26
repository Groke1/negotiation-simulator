import { ClusterBars } from './UI';

export default function ProfileMetricsCard({ clusters = [], dictionary = [], title, action, className = '' }) {
  return <section className={`card profile-card ${className}`.trim()}>
    {title && <div className="profile-card-heading"><h2>{title}</h2></div>}
    <ClusterBars clusters={clusters} dictionary={dictionary} variant="profile-clusters" />
    {action && <div className="profile-card-action">{action}</div>}
  </section>;
}
