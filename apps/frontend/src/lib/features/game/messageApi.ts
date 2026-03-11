// src/lib/features/game/messageApi.ts
import client from '$lib/api/client';
import type { MessageDTO, CreateMessageRequest } from './types';

export const messageApi = {
  // POST /matches/:matchId/messages - Send user message and receive AI response
  sendMessage: (matchId: string, req: CreateMessageRequest) =>
    client.post<MessageDTO>(`/api/matches/${matchId}/messages`, req),

  // GET /matches/:matchId/messages - Fetch conversation history
  getHistory: (matchId: string) =>
    client.get<MessageDTO[]>(`/api/matches/${matchId}/messages`)
};
