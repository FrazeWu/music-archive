# Agents Configuration for All Kanye Song Collection Website

## Overview
This document defines AI agents and tools to assist in building a Vue 3 website for collecting all Kanye West songs, featuring authentication, uploads, corrections, and a minimalist UI, with a clear definition of user roles and permissions.

## Database Schema (Updated 2026-01-18)

### Current Architecture
The system uses an optimized schema with many-to-many relationships and typed correction tables.

**Core Tables:**
- `Users` - User accounts with role-based access
- `Artists` - Artist information
- `Albums` - Album metadata with status tracking
- `Songs` - Song metadata with status tracking

**Junction Tables (Many-to-Many):**
- `album_artists` - Albums can have multiple artists (collaboration support)
- `song_artists` - Songs can have multiple artists (featured artists support)

**Correction Tables (Typed):**
- `song_corrections` - Structured song corrections (field-specific)
- `album_corrections` - Structured album corrections (title, cover, release date, artists)

**Key Features:**
- File source tracking (`audio_source`, `cover_source`: 'local' or 's3')
- Audit trail (ApprovedBy, RejectedBy, ApprovedAt, RejectedAt)
- Status management ('pending', 'approved', 'rejected')

## User Roles and Permissions

To ensure content quality and proper management, the system implements a user role and permission mechanism.

*   **Anonymous Users**:
    *   **Allowed Actions**: Browse all approved songs and albums, play music.
    *   **Restrictions**: Cannot upload, modify content, submit corrections, or access any authenticated features.

*   **Regular Users (`user`)**:
    *   **Authentication**: Register and login to receive JWT.
    *   **Allowed Actions**: Browse all approved songs/albums, play music, manage personal profile.
    *   **Actions Requiring Approval**: Submit new song/album uploads, submit correction requests for existing content. These actions enter a "pending" status and require admin approval before becoming visible/effective.
    *   **Restrictions**: Cannot directly edit or delete public content, no access to admin features.

*   **Admin Users (`admin`)**:
    *   **Authentication**: Login to receive JWT with admin role information.
    *   **Allowed Actions**: All permissions of regular users.
    *   **Immediate Effect Actions**:
        *   **Direct creation, editing, and deletion of any songs/albums**: Admin operations do not require approval and take effect immediately.
        *   **Direct submission of uploads and corrections**: Content submitted by admins via upload or correction interfaces is marked as "approved" and effective immediately, bypassing the "pending" workflow.
    *   **Core Responsibilities**: Review and approve/reject all song/album uploads and correction requests submitted by regular users.
    *   **Management Features**: User management (role changes, account suspension), system configuration.

## Agents

### Frontend Agent
- **Purpose**: Develop Vue 3 components, implement black/white minimalist UI with topbar, timeline, and player, adapting UI/UX based on user roles and content approval status.
- **Tools**: Vue Router, Pinia for state management, Tailwind CSS for styling, fetch API for HTTP calls.
- **Responsibilities**:
    *   Build `HomeView`, `AlbumDetailView`, `SongDetailView`, `UploadView`, `AdminReviewView`, `LoginView`, `EditAlbumView`, `AboutView`.
    *   Implement `AppTopbar` and `AudioPlayer` components.
    *   **Permissions and Roles**:
        *   Ensure `HomeView`, `AlbumDetailView`, `SongDetailView`, and `AudioPlayer` are accessible to anonymous users, displaying only `approved` content.
        *   Dynamically show/hide UI elements based on user role (anonymous, regular user, admin), such as upload buttons, edit/delete buttons, admin panel entry points.
        *   For regular users: After submitting uploads or corrections, clearly display "pending approval" status in the UI.
        *   For admins: After submitting uploads or making modifications, the UI should reflect that changes are effective immediately, without showing "pending approval" status.
        *   Build `AdminReviewView` to display all pending uploads and correction requests submitted by regular users, providing approve/reject interaction functionality, and showing the submitting user and modified information.
    *   Integrate user authentication status and role information into Pinia store for global access.
    *   **Type Definitions**: Use TypeScript interfaces for `Artist`, `Album`, `Song`, `SongCorrection`, `AlbumCorrection` matching backend schema.

### Backend Agent
- **Purpose**: Create API endpoints for user auth, song/album management, media upload, corrections, and admin review, with robust role-based access control.
- **Tools**: Go (Gin framework), GORM for database interaction, SQLite (development) / MySQL (production) database, JWT for authentication, AWS SDK for S3-compatible cloud storage.
- **Responsibilities**:
    *   Implement auth routes (register, login, token validation).
    *   Implement CRUD operations for songs and albums.
    *   Integrate cloud storage for media file uploads (AWS S3-compatible).
    *   Implement correction submission and application logic.
    *   **User Management**:
        *   Add `role` field to `User` model (`user`, `admin`), include role information in JWT upon login.
        *   `create_admin` utility tool for creating admin users.
    *   **Permissions and Access Control**:
        *   Set endpoints like `GET /api/albums`, `GET /api/albums/:id`, `GET /api/songs` (non-admin lists) as publicly accessible, but returning only `approved` status content.
        *   Implement middleware to verify permissions for protected API routes based on role information in JWT.
    *   **Content Approval and Status Management**:
        *   Add `status` field to `Song`, `Album`, `SongCorrection`, `AlbumCorrection` models (`pending`, `approved`, `rejected`).
        *   **Regular User Operations**: When submitting uploads or corrections, content is set to `pending` status by default, and immediately updated in the database and S3.
        *   **Admin Operations**: Content submitted by admins via upload, correction, or direct modification interfaces is set to `approved` status and takes effect immediately.
        *   Provide admin-only API endpoints:
            *   Retrieve all pending uploads and correction requests submitted by regular users.
            *   Approve or reject pending uploads and corrections. If approved, update the status to 'approved' and apply changes (for corrections). If rejected, permanently delete the content from the database and associated files from S3.
    *   **Correction Endpoints**:
        *   `POST /api/corrections/song` - Submit song corrections
        *   `POST /api/corrections/album` - Submit album corrections
        *   `GET /api/admin/pending-song-corrections` - List pending song corrections
        *   `POST /api/admin/approve-song-correction/:id` - Approve song correction
        *   `POST /api/admin/reject-song-correction/:id` - Reject song correction
        *   `GET /api/admin/pending-album-corrections` - List pending album corrections
        *   `POST /api/admin/approve-album-correction/:id` - Approve album correction
        *   `POST /api/admin/reject-album-correction/:id` - Reject album correction

### Data Agent
- **Purpose**: Collect and structure Kanye discography data.
- **Tools**: Web scraping (Puppeteer), APIs (Spotify, Genius), CSV/JSON processing.
- **Responsibilities**: Scrape song metadata, populate initial database.

### Media Agent
- **Purpose**: Handle media storage, streaming, and optimization.
- **Tools**: AWS S3 SDK (via Backend Agent), CloudFront CDN, FFmpeg for audio processing (may be an external service or integrated tool).
- **Responsibilities**: Upload files to cloud, generate streaming URLs, ensure secure access, optimize audio for playback.

### Testing Agent
- **Purpose**: Ensure code quality and functionality.
- **Tools**: Vitest for frontend unit tests, Playwright for frontend E2E tests, Go's testing framework for backend unit/integration tests.
- **Responsibilities**: Write tests for auth, uploads, playback, UI interactions, and role-based access control.

## Migration History

### 2026-01-18: Schema Optimization
**Changes Made:**
1. Album-Artist relationship changed from one-to-many to many-to-many (supports collaboration albums)
2. Correction table split into `SongCorrection` and `AlbumCorrection` with typed fields
3. Added file source tracking (`audio_source`, `cover_source`) for local staging support
4. Added audit trail fields to correction tables

**Files Modified:**
- Backend: `models.go`, `main.go`, `songs.go`, `albums.go`, `admin.go`, `corrections.go`
- Frontend: `types.ts`, `AdminReviewView.vue`

**Admin Credentials:**
- Username: `admin`
- Email: `admin@kanyearchive.com`
- Password: `admin123`

## Usage
Agents work collaboratively: Frontend builds UI, Backend handles logic, Data populates content, Media manages files, Testing validates.
