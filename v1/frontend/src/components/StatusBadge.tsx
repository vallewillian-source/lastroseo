import { StatusLight } from '@adobe/react-spectrum';

interface Props {
  status: string;
}
export default function StatusBadge({ status }: Props) {
  const variant = status === 'ok' || status === 'ready' || status === 'PONG'
    ? 'positive'
    : status === 'PENDING' || status === 'PROCESSING'
      ? 'notice'
      : 'negative';
  return <StatusLight variant={variant}>{status}</StatusLight>;
}
