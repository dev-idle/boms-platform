import { AdminUserDetail } from "@/features/admin";

type AdminUserDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function AdminUserDetailPage({
  params,
}: AdminUserDetailPageProps) {
  const { id } = await params;
  return <AdminUserDetail userId={id} />;
}
