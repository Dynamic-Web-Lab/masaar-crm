# Masaar CRM Training Guide

**Comprehensive documentation for end-users, administrators, and development teams.**

---

## Table of Contents

1. [User Guide](#user-guide)
2. [Administrator Guide](#administrator-guide)
3. [Feature Documentation](#feature-documentation)
4. [Workflow Tutorials](#workflow-tutorials)
5. [Troubleshooting](#troubleshooting)

---

## User Guide

### Getting Started

#### Login
1. Navigate to `http://your-masaar-instance.com`
2. Enter your email and password
3. Select your preferred language (Arabic or English)
4. Click **Sign In**

**Default Credentials** (development only):
- Email: `admin@masaar.local`
- Password: `changeme`

⚠️ **Important**: Change your password immediately after first login!

#### Dashboard Overview
- **Left Sidebar**: Navigation menu with all modules
- **Top Bar**: User profile, language selector, logout
- **Main Area**: Current page content
- **Notifications**: Real-time updates via WebSocket

### Core Features

#### Contact Management
**Purpose**: Unified contact database linked to WhatsApp conversations

**How to Add a Contact**:
1. Navigate to **Contacts** in the sidebar
2. Click **+ New Contact**
3. Enter required fields:
   - Phone (WhatsApp number with country code, e.g., +971XXXXXXXXX)
   - Full Name
   - Email (optional)
   - Preferred Language (Arabic/English)
4. Click **Save**

**How to View Contact Details**:
1. Click any contact name in the list
2. View all associated WhatsApp threads and communications
3. See assigned agent and lead score

#### Lead Management
**Purpose**: Sales pipeline tracking with Kanban board

**How to Create a Lead**:
1. Navigate to **Leads** in the sidebar
2. Click **+ New Lead** or link to existing contact
3. Set initial stage: **New** → **Contacted** → **Qualified** → **Proposal** → **Won/Lost**
4. Assign deal value and currency (AED recommended)
5. Add notes and tags
6. Click **Save**

**How to Move a Lead Through Stages**:
1. Open the **Kanban board** (default view in Leads)
2. Drag lead card to new column (stage)
3. System automatically updates lead status
4. Notifications sent to assigned agent

#### WhatsApp Communication
**Purpose**: Integrated WhatsApp messaging for seamless customer interaction

**How to Send a WhatsApp Message**:
1. Open a **Contact** or **Lead**
2. Scroll to **WhatsApp Thread** section
3. Type message in input field
4. Click **Send**
5. Message appears in real-time conversation

**Message Tagging** (AI-powered):
- Messages containing keywords automatically tagged
- Example: "interested" → tagged as "Interested"
- Helps with lead scoring and segmentation

#### Deal Tracking
**Purpose**: Monitor sales deals from opportunity to closure

**How to Create a Deal**:
1. Navigate to **Deals** in the sidebar
2. Click **+ New Deal**
3. Link to existing lead or create new contact
4. Enter deal value, currency, and stage
5. Assign to agent
6. Click **Save**

**Deal Stages**:
- **Negotiation**: Initial discussion
- **Proposal**: Sent proposal to client
- **Committed**: Client committed to purchase
- **Won**: Deal closed successfully
- **Lost**: Deal unsuccessful

#### Invoice Management
**Purpose**: Generate and track invoices with automatic VAT calculation

**How to Create an Invoice**:
1. Navigate to **Invoices** in the sidebar
2. Click **+ New Invoice**
3. Select customer/contact
4. Add line items (description, quantity, unit price)
5. System auto-calculates **5% UAE VAT**
6. Review total and click **Generate PDF**
7. Click **Send** to email customer

**Invoice Status**:
- **Draft**: Not yet sent
- **Sent**: Sent to customer
- **Paid**: Payment received

### Advanced Features

#### Analytics Dashboard
**Purpose**: Portfolio-level insights on business performance

**Overview Tab**:
- Total Tenants: Number of active tenants
- Occupancy Rate %: Units occupied / total units
- Total Monthly Revenue: Across all properties
- Collection Rate %: Rent collected / due rent

**Properties Tab**:
- View performance of each property
- Compare occupancy, revenue, and maintenance needs
- Sort by any column

**Tenants Tab**:
- Risk scoring (Red/Yellow/Green)
- Payment history and disputes
- Average stay duration
- Filter by risk level

**Financial Tab**:
- Revenue vs expenses over time
- Profit margins and net operating income
- Expense breakdown by category
- Date range filtering

#### Expense Tracking
**Purpose**: Record and manage operational expenses

**How to Log an Expense**:
1. Navigate to **Expenses** in the sidebar
2. Click **+ New Expense**
3. Select category (maintenance, utilities, etc.)
4. Enter amount and payment method (cash, transfer, card, check)
5. Upload receipt (optional but recommended)
6. Add vendor information
7. Click **Save**

**Expense Status**:
- **Pending**: Awaiting approval
- **Approved**: Approved by admin
- **Paid**: Payment made
- **Refunded**: Refund processed

**How to Request Approval**:
1. Open expense record
2. Click **Request Approval**
3. Add comments (optional)
4. Admin notified in real-time

#### Inspections
**Purpose**: Schedule property inspections and document findings

**How to Schedule an Inspection**:
1. Navigate to **Inspections** in the sidebar
2. Click **+ Schedule Inspection**
3. Select property and template (if available)
4. Set scheduled date and time
5. Assign inspector (optional)
6. Click **Save**

**How to Complete an Inspection**:
1. Open inspection record
2. Mark each checklist item: **Pass** or **Fail**
3. Add notes for failed items
4. Upload photos from inspection
5. Set severity level (Green/Yellow/Red)
6. Click **Complete**

#### Maintenance Tasks
**Purpose**: Track and manage maintenance work

**How to Create a Maintenance Task**:
1. Navigate to **Maintenance** in the sidebar
2. Click **+ New Task**
3. Select property and maintenance type (plumbing, electrical, HVAC, etc.)
4. Set priority (Low/Medium/High/Urgent)
5. Assign contractor or internal staff
6. Set due date
7. Estimate cost
8. Click **Save**

**How to Track Completion**:
1. Update task status: Pending → Scheduled → In Progress → Completed
2. Upload before/during/after photos
3. Enter actual cost when complete
4. Add completion notes

---

## Administrator Guide

### User Management
**Purpose**: Control system access and role permissions

**How to Add a User**:
1. Navigate to **Settings** → **Users**
2. Click **+ Add User**
3. Enter email, password, and full name
4. Assign role:
   - **Admin**: Full system access, manage users/settings
   - **Agent**: Create/update leads, contacts, deals
   - **Viewer**: Read-only access
5. Select preferred language
6. Click **Save**

⚠️ **Important**: Send users temporary password via secure channel. Instruct them to change on first login.

**How to Reset a User's Password**:
1. Navigate to **Settings** → **Users**
2. Click user name
3. Click **Reset Password**
4. Share new temporary password securely
5. User must change on next login

### Settings & Configuration

#### API Settings
- **Location**: **Settings** → **API**
- Manage WhatsApp webhook credentials
- Configure SMTP for email sending
- Set up integrations (optional)

#### Email Configuration
1. Navigate to **Settings** → **Email**
2. Enter SMTP host, port, username, password
3. Set from email and display name
4. Click **Test** to verify connection
5. Click **Save**

#### WhatsApp Setup
1. Navigate to **Settings** → **WhatsApp**
2. Enter Phone Number ID (from Meta/WhatsApp Business)
3. Enter Access Token
4. Enter Webhook Verify Token
5. Configure webhook URL in Meta Business Manager
6. Click **Verify**

#### Expense Category Management
1. Navigate to **Settings** → **Expense Categories**
2. Click **+ New Category**
3. Enter category name and type
4. Add description (optional)
5. Click **Save**

### Approval Workflows

#### Approving Expenses
1. Notifications alert admin when expense approval requested
2. Navigate to **Expenses** → filter by **Pending**
3. Review expense details (receipt, vendor, amount)
4. Click **Approve** or **Reject**
5. Add comments if rejecting
6. System notifies requester

#### Approving Commission Payouts
1. Navigate to **Commissions** (Phase 16.2)
2. Filter by **Pending**
3. Review agent performance: deals, leases, total commission
4. Click **Approve** to accept commission
5. Click **Mark as Paid** with payment reference
6. Agent receives payment notification

### Reporting

#### Run a Report
1. Navigate to **Reports** section
2. Select report type (financial, tenant, property, commission)
3. Set filters:
   - Date range
   - Property/agent filter
   - Status filter
4. Click **Generate**
5. Preview or download as CSV/PDF

#### Audit Log
1. Navigate to **Settings** → **Audit Log**
2. View all system actions with:
   - Timestamp
   - User who performed action
   - What changed (before/after values)
3. Filter by user, date range, action type
4. **Compliance**: Export log for regulatory review

### Backup & Recovery

**Automated Backups**:
- System automatically backs up PostgreSQL database daily
- Backups retained for 30 days
- Location: `/backups/` in Docker volume

**Manual Backup**:
```bash
docker compose exec postgres pg_dump -U masaar masaar > backup_$(date +%Y%m%d).sql
```

**Restore from Backup**:
```bash
docker compose exec -T postgres psql -U masaar masaar < backup_20260426.sql
```

---

## Feature Documentation

### Lease Renewal Automation (Phase 16.1)

**Purpose**: Automate tenant communication and lease renewal workflows

**How to Initiate a Renewal**:
1. Navigate to **Leases**
2. Click lease approaching expiration
3. Click **Initiate Renewal**
4. System creates renewal workflow 90 days before expiry
5. Select communication template (email/WhatsApp)
6. Click **Send Offer**

**Renewal Status Tracking**:
- **Pending**: Renewal created, awaiting communication
- **Offer Sent**: Renewal offer sent to tenant
- **Accepted**: Tenant accepted renewal terms
- **Counter-Offer**: Tenant submitted counter-offer
- **Rejected**: Tenant rejected renewal

**Managing Counter-Offers**:
1. Open renewal record
2. Review tenant's proposed rent amount
3. Accept, reject, or counter-propose
4. Add internal notes for approval
5. Notify tenant of decision

### Commission Tracking (Phase 16.2)

**Purpose**: Calculate and manage agent commissions

**Commission Structure Setup**:
1. Navigate to **Settings** → **Commission Structures**
2. Create structure with:
   - Commission type: Fixed amount, Percentage, or Tiered
   - Applicable to: Deals, Leases, or Both
   - Rules: Define how commission calculated
3. Set effective date range
4. Click **Save**

**Monthly Commission Generation**:
- System automatically calculates commissions on 1st of month
- Pulls all closed deals/leases from previous month
- Applies configured commission structure
- Creates commission record with status **Pending Approval**

**Commission Approval**:
1. Navigate to **Commissions** → **Pending**
2. Review agent performance details
3. Click **Approve** or **Dispute**
4. Click **Mark as Paid** with payment reference
5. Agent notified of payment

### Document Management (Phase 16.3)

**Purpose**: Centralize and manage legally-binding documents

**Creating a Document Template**:
1. Navigate to **Settings** → **Document Templates**
2. Click **+ New Template**
3. Select document type (lease, offer, inspection, waiver, custom)
4. Enter template name and language
5. Paste template content with placeholders: `{{tenant_name}}`, `{{property_address}}`, etc.
6. Enable signature requirement (optional)
7. Click **Save**

**Generating a Document from Template**:
1. Navigate to **Documents** → **+ New Document**
2. Select template
3. Auto-fills related entity (property, tenant, lease)
4. Review field values
5. Click **Generate PDF**
6. Click **Send** to email tenant

**Signature Workflow**:
1. Recipient receives document via email with signature link
2. Opens link and enters signature (digital signature pad)
3. System records timestamp, IP address, user-agent
4. Document marked as **Signed**
5. Email confirmation sent to both parties

---

## Workflow Tutorials

### Complete Lead-to-Invoice Workflow

**Step 1: Receive WhatsApp Inquiry**
1. Tenant inquires about property via WhatsApp
2. System creates/links contact automatically
3. Agent receives notification

**Step 2: Create Lead**
1. Agent views contact details
2. Clicks **Create Lead** from contact
3. Sets initial stage: **New**
4. Adds property of interest
5. Saves lead

**Step 3: Communicate & Qualify**
1. Agent exchanges messages via WhatsApp
2. Tags important messages (e.g., "Interested", "Budget")
3. AI automatically scores lead based on:
   - Message content
   - Engagement frequency
   - Tags
4. Lead moves to **Contacted** → **Qualified**

**Step 4: Create Deal**
1. Agent confident tenant will sign lease
2. Creates **Deal** linked to lead
3. Sets expected monthly rent
4. Sets deal stage to **Proposal**

**Step 5: Generate and Send Invoice**
1. Agreement reached on rent and terms
2. Navigate to **Invoices**
3. Create invoice for deposit/first month
4. System auto-adds 5% UAE VAT
5. Click **Generate PDF**
6. Click **Send** → invoice emailed with payment link
7. Tenant pays online
8. System marks invoice **Paid**

**Step 6: Track Lease**
1. Lease recorded with all terms
2. Monthly reminders for rent collection
3. Analytics updated with revenue
4. Renewal workflow auto-starts 90 days before expiry

### Property Inspection Workflow

**Step 1: Schedule Inspection**
1. Pre-lease inspection scheduled
2. Select inspection template (e.g., "Pre-Lease Standard")
3. Set inspection date and assign inspector

**Step 2: Conduct Inspection**
1. Inspector arrives at property
2. Opens inspection record on mobile
3. Goes through checklist items (walls, plumbing, HVAC, etc.)
4. Marks each item: **Pass** or **Fail**
5. Takes photos for failed items

**Step 3: Document Findings**
1. Inspector sets overall severity: **Green** (good) / **Yellow** (minor issues) / **Red** (critical)
2. Adds detailed notes for any red flags
3. Uploads photos (before if existing damage)
4. Clicks **Complete Inspection**

**Step 4: Create Maintenance Tasks**
1. System suggests maintenance tasks based on inspection findings
2. Inspector can create tasks directly from inspection
3. Each task assigned priority (Low/Medium/High/Urgent)
4. Estimated cost added
5. Contractor assigned

**Step 5: Track Maintenance**
1. Inspector schedules maintenance dates
2. Takes before photos
3. Contractor completes work and uploads after photos
4. Actual cost recorded
5. Task marked **Completed**

### Commission Calculation Workflow

**Step 1: Deal Closed**
1. Agent closes deal (e.g., AED 100,000 property rental annually)
2. System records in **Deals** with status **Won**
3. Lease created with signed date = current month

**Step 2: Monthly Commission Calculation**
1. On 1st of next month, system runs background job
2. Queries all deals closed previous month
3. Applies commission structure for agent:
   - Example: 5% of deal value
   - Calculates: 100,000 × 5% = AED 5,000
4. Creates `AgentCommission` record with status **Pending**
5. Admin receives notification

**Step 3: Approve & Pay**
1. Admin navigates to **Commissions** → **Pending**
2. Reviews: Agent name, period, deals count, total commission
3. Clicks **Approve**
4. Commission moves to **Approved**
5. Admin clicks **Mark as Paid**, enters bank transfer reference
6. Agent notified of payment

---

## Troubleshooting

### Common Issues

#### "Unauthorized" Error on Login
**Cause**: Invalid credentials or token expired

**Solution**:
1. Clear browser cookies: **Settings** → **Clear Site Data**
2. Refresh page (Ctrl+R or Cmd+R)
3. Re-enter email and password
4. If still fails, request password reset from admin

#### WhatsApp Messages Not Showing
**Cause**: Webhook not configured or token expired

**Solution**:
1. Admin navigates to **Settings** → **WhatsApp**
2. Verify Access Token is current (Meta rotates annually)
3. Verify webhook URL is correct in Meta Business Manager
4. Check server logs: `docker compose logs server`
5. Look for "webhook verification failed" errors

#### Expense Approval Notification Not Received
**Cause**: WebSocket disconnected or notification disabled

**Solution**:
1. Refresh page to re-establish WebSocket
2. Check browser console for errors: Press F12 → **Console** tab
3. Verify user has **Admin** role
4. Verify expense requester selected correct approver

#### Slow Analytics Loading
**Cause**: Large dataset requiring complex aggregations

**Solution**:
1. Filter by date range to reduce dataset
2. Use property filter to focus on specific properties
3. Check server logs for slow query warnings
4. If persistent, contact admin for database optimization

### Getting Help

**For Users**:
- Click **?** (help icon) in top-right corner
- Review this **TRAINING.md** document
- Contact your administrator

**For Administrators**:
1. Check **Settings** → **Audit Log** for any failed operations
2. Review server logs: `docker compose logs server | tail -50`
3. Check database connectivity: `docker compose ps` (all should be running)
4. Restart server if needed: `docker compose restart server`

**For Developers**:
- See **CLAUDE.md** for architecture and API documentation
- Check **PHASE-15-16-ROADMAP.md** for planned features
- Review commit history: `git log --oneline`

---

## Security Best Practices

### For Users
1. **Change default password** immediately after first login
2. **Never share** your login credentials
3. **Enable 2FA** if available (future release)
4. **Log out** when leaving your desk
5. **Report suspicious activity** to admin immediately

### For Administrators
1. **Rotate** WhatsApp access tokens annually
2. **Backup** database daily (automated in Docker)
3. **Update** system regularly: `git pull && docker compose pull && docker compose up`
4. **Monitor** audit log for unauthorized access
5. **Review** user permissions quarterly

### For Developers
1. **Use HTTPS** in production (configure via reverse proxy)
2. **Set strong JWT secret**: At least 32 random characters
3. **Rotate** database passwords quarterly
4. **Restrict** Redis access: Use `requirepass` setting
5. **Keep dependencies** updated: `go get -u ./...`

---

**Version**: 0.2.0 (Phases 15-16)
**Last Updated**: April 2026
**Maintainer**: Dori Internet Dev Team
**Next Review**: Phase 17 launch (estimated Q3 2026)
