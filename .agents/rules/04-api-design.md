---
trigger: always_on
---

# RESTful API Design Guidelines
1. **Resource-Oriented URLs**: URI는 행위(동사)가 아닌 **명사(복수형)**를 사용하세요.
   - ⭕ 좋은 예: `GET /users`, `POST /matches/{id}/join`
   - ❌ 나쁜 예: `POST /createUser`, `GET /getMatch`
2. **HTTP Methods**: 용도에 맞는 정확한 메서드를 사용하세요.
   - `GET`: 리소스 조회 (응답: 200 OK)
   - `POST`: 리소스 생성 (응답: 201 Created)
   - `PUT`: 리소스 전체 교체 (응답: 200 OK 또는 204 No Content)
   - `PATCH`: 리소스 부분 수정 (응답: 200 OK)
   - `DELETE`: 리소스 삭제 (응답: 200 OK 또는 204 No Content)
3. **Query Parameters**: 필터링, 정렬, 페이징 등은 경로(Path)가 아닌 쿼리 스트링을 사용하세요.
   - 예: `GET /games?status=active&page=1&limit=20`
4. **HTTP Status Codes**: 상황에 맞는 상태 코드를 엄격히 반환하세요.
   - `400 Bad Request`: 파라미터 검증 실패, 잘못된 요청 포맷
   - `401 Unauthorized`: 토큰 없음, 유효하지 않은 토큰
   - `403 Forbidden`: 토큰은 있으나 해당 액션에 대한 권한(Role) 부족
   - `404 Not Found`: 요청한 리소스가 존재하지 않음
5. **Error Handling**: 에러 처리에 대해서는 반드시 클라이언트에게 domain에서 정의된 error를 리턴하도록 하세요.

# Handler & Auth Context Guidelines
1. **User ID 추출 (JWT)**:
   - 핸들러의 컨텍스트에서 `user_id`를 안전하게 타입 캐스팅하여 가져오세요.
2. **Insecure Direct Object Reference (IDOR) 취약점 방지**:
   - 클라이언트가 전달한 경로 파라미터나 요청 본문의 리소스 ID만 믿고 데이터를 조작하거나 반환하지 마세요.
   - 리소스(예: 유저 프로필, 게임 매치 데이터 등)에 접근하거나 수정할 때, **항상 DB에서 조회한 리소스의 실제 소유자 ID와 JWT에서 추출한 현재 로그인한 `user_id`가 일치하는지 검증**해야 합니다.
   - 이 소유권 검증 로직은 Handler가 아닌 UseCase에서 수행하세요.
3. **Response Format**: 
   - 성공 응답과 에러 응답의 JSON 구조를 일관되게 유지하세요.