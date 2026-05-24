import { AdminUserDetail } from "@/features/admin";

export default function AdminUserDetailPage({
  params,
}: {
  params: { id: string };
}) {
  return <AdminUserDetail userId={params.id} />;
}
