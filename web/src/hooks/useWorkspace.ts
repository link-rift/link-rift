import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { useAuthStore } from "@/stores/authStore"
import * as workspaceService from "@/services/workspaces"
import type {
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  InviteMemberRequest,
  UpdateMemberRoleRequest,
  TransferOwnershipRequest,
} from "@/types/workspace"

export function useWorkspaces() {
  const { setWorkspaces, clearWorkspaces } = useWorkspaceStore()
  const { isAuthenticated } = useAuthStore()

  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      try {
        const workspaces = await workspaceService.getWorkspaces()
        setWorkspaces(workspaces)
        return workspaces
      } catch {
        clearWorkspaces()
        return []
      }
    },
    enabled: isAuthenticated,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })
}

export function useWorkspace(id: string) {
  return useQuery({
    queryKey: ["workspaces", id],
    queryFn: () => workspaceService.getWorkspace(id),
    enabled: !!id,
  })
}

export function useCreateWorkspace() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateWorkspaceRequest) => workspaceService.createWorkspace(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] })
    },
  })
}

export function useUpdateWorkspace() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateWorkspaceRequest }) =>
      workspaceService.updateWorkspace(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] })
    },
  })
}

export function useDeleteWorkspace() {
  const { clearWorkspaces } = useWorkspaceStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => workspaceService.deleteWorkspace(id),
    onSuccess: () => {
      clearWorkspaces()
      queryClient.invalidateQueries({ queryKey: ["workspaces"] })
    },
  })
}

export function useWorkspaceMembers(workspaceId: string) {
  return useQuery({
    queryKey: ["workspaces", workspaceId, "members"],
    queryFn: () => workspaceService.getMembers(workspaceId),
    enabled: !!workspaceId,
    staleTime: 30 * 1000,
  })
}

export function useInviteMember() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ workspaceId, data }: { workspaceId: string; data: InviteMemberRequest }) =>
      workspaceService.inviteMember(workspaceId, data),
    onSuccess: (_, { workspaceId }) => {
      queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "members"] })
      queryClient.invalidateQueries({ queryKey: ["workspaces"] })
    },
  })
}

export function useUpdateMemberRole() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ workspaceId, userId, data }: { workspaceId: string; userId: string; data: UpdateMemberRoleRequest }) =>
      workspaceService.updateMemberRole(workspaceId, userId, data),
    onSuccess: (_, { workspaceId }) => {
      queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "members"] })
    },
  })
}

export function useRemoveMember() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ workspaceId, userId }: { workspaceId: string; userId: string }) =>
      workspaceService.removeMember(workspaceId, userId),
    onSuccess: (_, { workspaceId }) => {
      queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "members"] })
      queryClient.invalidateQueries({ queryKey: ["workspaces"] })
    },
  })
}

export function useTransferOwnership() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ workspaceId, data }: { workspaceId: string; data: TransferOwnershipRequest }) =>
      workspaceService.transferOwnership(workspaceId, data),
    onSuccess: (_, { workspaceId }) => {
      queryClient.invalidateQueries({ queryKey: ["workspaces", workspaceId, "members"] })
      queryClient.invalidateQueries({ queryKey: ["workspaces"] })
    },
  })
}
