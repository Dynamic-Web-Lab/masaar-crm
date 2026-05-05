# Masaar-CRM — Claude Design Context

## 🎯 Project Goal
Bilingual (Arabic/English) WhatsApp-first CRM for UAE businesses. PDPL-compliant, self-hosted.

## 🎨 Design System
- **Framework**: Next.js 14 App Router + Tailwind CSS
- **Typography**: 
  - English: `Inter` (font-sans)
  - Arabic: `Cairo` (font-arabic) with `dir="rtl"`
- **Colors**: 
  - Primary: `#2563eb` (blue-600)
  - Accent: `#d97706` (amber-600, UAE gold)
  - Background: `#f8fafc` (slate-50) / Dark: `#0f172a`
- **Spacing**: 4px baseline grid (Tailwind default)
- **Components**: `/web/components/ui/` (Radix-style primitives)

## 🌐 RTL/LTR Handling
- Language context: `/web/context/LanguageContext.tsx`
- CSS: Use logical properties (`ms-`, `me-`, `start-`, `end-`) NOT `left`/`right`
- Dir attribute: `<html dir={lang === 'ar' ? 'rtl' : 'ltr'}>`

## 📱 Responsive Breakpoints
- Mobile: `<640px` (sales teams use phones for WhatsApp)
- Tablet: `640px–1024px`
- Desktop: `>1024px`

## ♿ Accessibility
- WCAG 2.1 AA target
- All interactive elements: keyboard navigable + ARIA labels in both languages

## 🔑 Key Pages to Design
1. `/dashboard` — Stats overview + quick actions
2. `/inbox` — WhatsApp thread list + message view
3. `/pipeline` — Kanban board for leads/deals
4. `/contacts` — List + detail modals
5. `/settings` — User profile + language toggle

## 🤖 AI Features (Local Ollama)
- Thread summarization button
- Lead scoring badges (High/Medium/Low)
- Smart reply suggestions (optional toggle)

## 🚫 Constraints
- No external CDNs (self-hosted requirement)
- All AI processing must stay on-prem (PDPL compliance)
- Arabic text must render correctly in PDF invoices