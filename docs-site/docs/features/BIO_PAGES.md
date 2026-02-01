# Bio Pages

> Last Updated: 2025-01-24

Linkrift provides a powerful link-in-bio page builder that enables users to create customizable landing pages showcasing multiple links, media, and content blocks with drag-and-drop functionality.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Theme System](#theme-system)
- [Drag-and-Drop with dnd-kit](#drag-and-drop-with-dnd-kit)
- [Custom CSS](#custom-css)
- [Media Embedding](#media-embedding)
- [SEO Optimization](#seo-optimization)
- [API Endpoints](#api-endpoints)
- [React Components](#react-components)

---

## Overview

Bio Pages in Linkrift offer:

- **Customizable landing pages** with personal branding
- **Multiple content blocks** including links, text, images, videos, and social embeds
- **Drag-and-drop editor** using dnd-kit for intuitive block arrangement
- **Theme system** with predefined and custom themes
- **Custom CSS** for advanced styling
- **SEO optimization** for better discoverability
- **Analytics** for page views and link clicks

## Architecture

```go
// internal/biopages/models.go
package biopages

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// BioPage represents a link-in-bio page
type BioPage struct {
	ID            string         `json:"id" db:"id"`
	WorkspaceID   string         `json:"workspace_id" db:"workspace_id"`
	Slug          string         `json:"slug" db:"slug"`
	Title         string         `json:"title" db:"title"`
	Description   string         `json:"description" db:"description"`
	AvatarURL     string         `json:"avatar_url" db:"avatar_url"`
	ThemeID       string         `json:"theme_id" db:"theme_id"`
	CustomCSS     string         `json:"custom_css,omitempty" db:"custom_css"`
	CustomDomain  string         `json:"custom_domain,omitempty" db:"custom_domain"`
	Blocks        Blocks         `json:"blocks" db:"blocks"`
	SEOSettings   SEOSettings    `json:"seo_settings" db:"seo_settings"`
	SocialLinks   []SocialLink   `json:"social_links" db:"social_links"`
	IsPublished   bool           `json:"is_published" db:"is_published"`
	TotalViews    int64          `json:"total_views" db:"total_views"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

// Block represents a content block on a bio page
type Block struct {
	ID        string          `json:"id"`
	Type      BlockType       `json:"type"`
	Content   json.RawMessage `json:"content"`
	IsVisible bool            `json:"is_visible"`
	Order     int             `json:"order"`
}

// BlockType defines the type of content block
type BlockType string

const (
	BlockTypeLink     BlockType = "link"
	BlockTypeHeader   BlockType = "header"
	BlockTypeText     BlockType = "text"
	BlockTypeImage    BlockType = "image"
	BlockTypeVideo    BlockType = "video"
	BlockTypeSocial   BlockType = "social"
	BlockTypeEmbed    BlockType = "embed"
	BlockTypeDivider  BlockType = "divider"
	BlockTypeSpotify  BlockType = "spotify"
	BlockTypeYouTube  BlockType = "youtube"
)

// Blocks is a slice of Block with JSON scanning
type Blocks []Block

func (b Blocks) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *Blocks) Scan(value interface{}) error {
	if value == nil {
		*b = []Block{}
		return nil
	}
	return json.Unmarshal(value.([]byte), b)
}

// LinkBlockContent represents link block content
type LinkBlockContent struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	LinkID      string `json:"link_id,omitempty"` // If using Linkrift shortened URL
	Style       string `json:"style"`             // default, featured, outline
}

// HeaderBlockContent represents header block content
type HeaderBlockContent struct {
	Text  string `json:"text"`
	Level int    `json:"level"` // 1, 2, 3
	Align string `json:"align"` // left, center, right
}

// TextBlockContent represents text block content
type TextBlockContent struct {
	Text  string `json:"text"`
	Align string `json:"align"`
}

// ImageBlockContent represents image block content
type ImageBlockContent struct {
	URL     string `json:"url"`
	Alt     string `json:"alt"`
	LinkURL string `json:"link_url,omitempty"`
	Width   string `json:"width"` // full, medium, small
}

// VideoBlockContent represents video block content
type VideoBlockContent struct {
	URL      string `json:"url"`
	Provider string `json:"provider"` // youtube, vimeo, etc.
	VideoID  string `json:"video_id"`
}

// EmbedBlockContent represents embed block content
type EmbedBlockContent struct {
	HTML   string `json:"html"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SocialLink represents a social media link
type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Icon     string `json:"icon"`
}

// SEOSettings contains SEO configuration
type SEOSettings struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Keywords        []string `json:"keywords"`
	OGImage         string   `json:"og_image"`
	TwitterCard     string   `json:"twitter_card"`
	CanonicalURL    string   `json:"canonical_url,omitempty"`
	NoIndex         bool     `json:"no_index"`
}
```

---

## Theme System

```go
// internal/biopages/themes.go
package biopages

// Theme represents a bio page theme
type Theme struct {
	ID          string           `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	PreviewURL  string           `json:"preview_url" db:"preview_url"`
	IsSystem    bool             `json:"is_system" db:"is_system"`
	IsPremium   bool             `json:"is_premium" db:"is_premium"`
	Styles      ThemeStyles      `json:"styles" db:"styles"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
}

// ThemeStyles defines the visual styling of a theme
type ThemeStyles struct {
	// Background
	BackgroundType  string `json:"background_type"` // solid, gradient, image
	BackgroundColor string `json:"background_color"`
	BackgroundGradient *GradientConfig `json:"background_gradient,omitempty"`
	BackgroundImage string `json:"background_image,omitempty"`

	// Typography
	FontFamily      string `json:"font_family"`
	FontURL         string `json:"font_url,omitempty"` // Google Fonts URL
	TitleColor      string `json:"title_color"`
	TextColor       string `json:"text_color"`
	LinkTextColor   string `json:"link_text_color"`

	// Links/Buttons
	ButtonStyle        string `json:"button_style"` // filled, outline, soft
	ButtonColor        string `json:"button_color"`
	ButtonTextColor    string `json:"button_text_color"`
	ButtonBorderRadius int    `json:"button_border_radius"`
	ButtonShadow       string `json:"button_shadow,omitempty"`
	ButtonHoverEffect  string `json:"button_hover_effect"` // none, lift, glow, scale

	// Layout
	ContentWidth     string `json:"content_width"` // narrow, medium, wide
	ContentAlignment string `json:"content_alignment"` // left, center
	BlockSpacing     int    `json:"block_spacing"`

	// Avatar
	AvatarSize        int    `json:"avatar_size"`
	AvatarBorderRadius int   `json:"avatar_border_radius"` // 0 = square, 50 = circle
	AvatarBorder      string `json:"avatar_border,omitempty"`

	// Social Icons
	SocialIconStyle string `json:"social_icon_style"` // filled, outline, minimal
	SocialIconColor string `json:"social_icon_color"`
	SocialIconSize  int    `json:"social_icon_size"`
}

// GradientConfig defines gradient settings
type GradientConfig struct {
	Type      string   `json:"type"` // linear, radial
	Angle     int      `json:"angle"`
	Colors    []string `json:"colors"`
	Positions []int    `json:"positions"` // Percentages
}

// PredefinedThemes contains built-in themes
var PredefinedThemes = map[string]Theme{
	"minimal_light": {
		ID:          "minimal_light",
		Name:        "Minimal Light",
		Description: "Clean and simple light theme",
		IsSystem:    true,
		Styles: ThemeStyles{
			BackgroundType:     "solid",
			BackgroundColor:    "#FFFFFF",
			FontFamily:         "Inter",
			TitleColor:         "#1a1a1a",
			TextColor:          "#4a4a4a",
			LinkTextColor:      "#1a1a1a",
			ButtonStyle:        "filled",
			ButtonColor:        "#1a1a1a",
			ButtonTextColor:    "#FFFFFF",
			ButtonBorderRadius: 8,
			ButtonHoverEffect:  "lift",
			ContentWidth:       "medium",
			ContentAlignment:   "center",
			BlockSpacing:       16,
			AvatarSize:         96,
			AvatarBorderRadius: 50,
			SocialIconStyle:    "minimal",
			SocialIconColor:    "#4a4a4a",
			SocialIconSize:     24,
		},
	},
	"minimal_dark": {
		ID:          "minimal_dark",
		Name:        "Minimal Dark",
		Description: "Sleek dark mode theme",
		IsSystem:    true,
		Styles: ThemeStyles{
			BackgroundType:     "solid",
			BackgroundColor:    "#0f0f0f",
			FontFamily:         "Inter",
			TitleColor:         "#FFFFFF",
			TextColor:          "#a0a0a0",
			LinkTextColor:      "#FFFFFF",
			ButtonStyle:        "filled",
			ButtonColor:        "#FFFFFF",
			ButtonTextColor:    "#0f0f0f",
			ButtonBorderRadius: 8,
			ButtonHoverEffect:  "glow",
			ContentWidth:       "medium",
			ContentAlignment:   "center",
			BlockSpacing:       16,
			AvatarSize:         96,
			AvatarBorderRadius: 50,
			SocialIconStyle:    "minimal",
			SocialIconColor:    "#FFFFFF",
			SocialIconSize:     24,
		},
	},
	"gradient_sunset": {
		ID:          "gradient_sunset",
		Name:        "Sunset Gradient",
		Description: "Warm gradient background",
		IsSystem:    true,
		IsPremium:   true,
		Styles: ThemeStyles{
			BackgroundType: "gradient",
			BackgroundGradient: &GradientConfig{
				Type:      "linear",
				Angle:     135,
				Colors:    []string{"#667eea", "#764ba2", "#f093fb"},
				Positions: []int{0, 50, 100},
			},
			FontFamily:         "Poppins",
			TitleColor:         "#FFFFFF",
			TextColor:          "#f0f0f0",
			LinkTextColor:      "#FFFFFF",
			ButtonStyle:        "soft",
			ButtonColor:        "rgba(255,255,255,0.2)",
			ButtonTextColor:    "#FFFFFF",
			ButtonBorderRadius: 12,
			ButtonHoverEffect:  "scale",
			ContentWidth:       "medium",
			ContentAlignment:   "center",
			BlockSpacing:       20,
			AvatarSize:         120,
			AvatarBorderRadius: 50,
			AvatarBorder:       "4px solid rgba(255,255,255,0.3)",
			SocialIconStyle:    "filled",
			SocialIconColor:    "#FFFFFF",
			SocialIconSize:     28,
		},
	},
}

// ThemeService handles theme operations
type ThemeService struct {
	repo *db.ThemeRepository
}

// GenerateCSS generates CSS from theme styles
func (ts *ThemeService) GenerateCSS(theme *Theme) string {
	styles := theme.Styles
	css := &strings.Builder{}

	// Font import
	if styles.FontURL != "" {
		css.WriteString(fmt.Sprintf("@import url('%s');\n\n", styles.FontURL))
	}

	// Root variables
	css.WriteString(":root {\n")
	css.WriteString(fmt.Sprintf("  --font-family: '%s', sans-serif;\n", styles.FontFamily))
	css.WriteString(fmt.Sprintf("  --title-color: %s;\n", styles.TitleColor))
	css.WriteString(fmt.Sprintf("  --text-color: %s;\n", styles.TextColor))
	css.WriteString(fmt.Sprintf("  --button-color: %s;\n", styles.ButtonColor))
	css.WriteString(fmt.Sprintf("  --button-text-color: %s;\n", styles.ButtonTextColor))
	css.WriteString(fmt.Sprintf("  --button-radius: %dpx;\n", styles.ButtonBorderRadius))
	css.WriteString(fmt.Sprintf("  --block-spacing: %dpx;\n", styles.BlockSpacing))
	css.WriteString("}\n\n")

	// Body styles
	css.WriteString(".bio-page {\n")
	css.WriteString(fmt.Sprintf("  font-family: var(--font-family);\n"))
	css.WriteString(fmt.Sprintf("  color: var(--text-color);\n"))

	switch styles.BackgroundType {
	case "solid":
		css.WriteString(fmt.Sprintf("  background-color: %s;\n", styles.BackgroundColor))
	case "gradient":
		if styles.BackgroundGradient != nil {
			g := styles.BackgroundGradient
			gradientCSS := ts.generateGradientCSS(g)
			css.WriteString(fmt.Sprintf("  background: %s;\n", gradientCSS))
		}
	case "image":
		css.WriteString(fmt.Sprintf("  background-image: url('%s');\n", styles.BackgroundImage))
		css.WriteString("  background-size: cover;\n")
		css.WriteString("  background-position: center;\n")
	}

	css.WriteString("}\n\n")

	// Content container
	widthMap := map[string]string{
		"narrow": "420px",
		"medium": "520px",
		"wide":   "680px",
	}
	css.WriteString(".bio-content {\n")
	css.WriteString(fmt.Sprintf("  max-width: %s;\n", widthMap[styles.ContentWidth]))
	css.WriteString(fmt.Sprintf("  text-align: %s;\n", styles.ContentAlignment))
	css.WriteString("}\n\n")

	// Button styles
	css.WriteString(".bio-link {\n")
	css.WriteString(fmt.Sprintf("  background-color: var(--button-color);\n"))
	css.WriteString(fmt.Sprintf("  color: var(--button-text-color);\n"))
	css.WriteString(fmt.Sprintf("  border-radius: var(--button-radius);\n"))

	if styles.ButtonStyle == "outline" {
		css.WriteString("  background-color: transparent;\n")
		css.WriteString(fmt.Sprintf("  border: 2px solid var(--button-color);\n"))
		css.WriteString(fmt.Sprintf("  color: var(--button-color);\n"))
	}

	if styles.ButtonShadow != "" {
		css.WriteString(fmt.Sprintf("  box-shadow: %s;\n", styles.ButtonShadow))
	}
	css.WriteString("}\n\n")

	// Hover effects
	switch styles.ButtonHoverEffect {
	case "lift":
		css.WriteString(".bio-link:hover {\n")
		css.WriteString("  transform: translateY(-4px);\n")
		css.WriteString("  box-shadow: 0 4px 12px rgba(0,0,0,0.15);\n")
		css.WriteString("}\n")
	case "glow":
		css.WriteString(".bio-link:hover {\n")
		css.WriteString(fmt.Sprintf("  box-shadow: 0 0 20px %s40;\n", styles.ButtonColor))
		css.WriteString("}\n")
	case "scale":
		css.WriteString(".bio-link:hover {\n")
		css.WriteString("  transform: scale(1.02);\n")
		css.WriteString("}\n")
	}

	// Avatar styles
	css.WriteString("\n.bio-avatar {\n")
	css.WriteString(fmt.Sprintf("  width: %dpx;\n", styles.AvatarSize))
	css.WriteString(fmt.Sprintf("  height: %dpx;\n", styles.AvatarSize))
	css.WriteString(fmt.Sprintf("  border-radius: %d%%;\n", styles.AvatarBorderRadius))
	if styles.AvatarBorder != "" {
		css.WriteString(fmt.Sprintf("  border: %s;\n", styles.AvatarBorder))
	}
	css.WriteString("}\n")

	return css.String()
}

func (ts *ThemeService) generateGradientCSS(g *GradientConfig) string {
	colorStops := make([]string, len(g.Colors))
	for i, color := range g.Colors {
		if i < len(g.Positions) {
			colorStops[i] = fmt.Sprintf("%s %d%%", color, g.Positions[i])
		} else {
			colorStops[i] = color
		}
	}

	if g.Type == "radial" {
		return fmt.Sprintf("radial-gradient(circle, %s)", strings.Join(colorStops, ", "))
	}
	return fmt.Sprintf("linear-gradient(%ddeg, %s)", g.Angle, strings.Join(colorStops, ", "))
}
```

---

## Drag-and-Drop with dnd-kit

```typescript
// src/components/biopages/BioPageEditor.tsx
import React, { useState, useCallback } from 'react';
import {
  DndContext,
  DragOverlay,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
  DragStartEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Block, BioPage } from '@/types/biopage';
import { BlockRenderer } from './BlockRenderer';
import { BlockEditor } from './BlockEditor';
import { AddBlockMenu } from './AddBlockMenu';

interface BioPageEditorProps {
  page: BioPage;
  onChange: (page: BioPage) => void;
}

export const BioPageEditor: React.FC<BioPageEditorProps> = ({ page, onChange }) => {
  const [activeId, setActiveId] = useState<string | null>(null);
  const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = page.blocks.findIndex((block) => block.id === active.id);
      const newIndex = page.blocks.findIndex((block) => block.id === over.id);

      const newBlocks = arrayMove(page.blocks, oldIndex, newIndex).map(
        (block, index) => ({ ...block, order: index })
      );

      onChange({ ...page, blocks: newBlocks });
    }

    setActiveId(null);
  };

  const handleBlockChange = useCallback(
    (blockId: string, content: any) => {
      const newBlocks = page.blocks.map((block) =>
        block.id === blockId ? { ...block, content } : block
      );
      onChange({ ...page, blocks: newBlocks });
    },
    [page, onChange]
  );

  const handleAddBlock = useCallback(
    (type: string) => {
      const newBlock: Block = {
        id: `block_${Date.now()}`,
        type,
        content: getDefaultContent(type),
        is_visible: true,
        order: page.blocks.length,
      };
      onChange({ ...page, blocks: [...page.blocks, newBlock] });
      setSelectedBlockId(newBlock.id);
    },
    [page, onChange]
  );

  const handleDeleteBlock = useCallback(
    (blockId: string) => {
      const newBlocks = page.blocks
        .filter((block) => block.id !== blockId)
        .map((block, index) => ({ ...block, order: index }));
      onChange({ ...page, blocks: newBlocks });
      if (selectedBlockId === blockId) {
        setSelectedBlockId(null);
      }
    },
    [page, onChange, selectedBlockId]
  );

  const handleToggleVisibility = useCallback(
    (blockId: string) => {
      const newBlocks = page.blocks.map((block) =>
        block.id === blockId ? { ...block, is_visible: !block.is_visible } : block
      );
      onChange({ ...page, blocks: newBlocks });
    },
    [page, onChange]
  );

  const activeBlock = activeId
    ? page.blocks.find((block) => block.id === activeId)
    : null;

  const selectedBlock = selectedBlockId
    ? page.blocks.find((block) => block.id === selectedBlockId)
    : null;

  return (
    <div className="flex h-full">
      {/* Editor Panel */}
      <div className="flex-1 overflow-auto p-6">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={page.blocks.map((b) => b.id)}
            strategy={verticalListSortingStrategy}
          >
            <div className="space-y-4 max-w-lg mx-auto">
              {page.blocks.map((block) => (
                <SortableBlock
                  key={block.id}
                  block={block}
                  isSelected={selectedBlockId === block.id}
                  onSelect={() => setSelectedBlockId(block.id)}
                  onDelete={() => handleDeleteBlock(block.id)}
                  onToggleVisibility={() => handleToggleVisibility(block.id)}
                />
              ))}
            </div>
          </SortableContext>

          <DragOverlay>
            {activeBlock ? (
              <div className="opacity-80">
                <BlockRenderer block={activeBlock} isPreview />
              </div>
            ) : null}
          </DragOverlay>
        </DndContext>

        <div className="max-w-lg mx-auto mt-6">
          <AddBlockMenu onAdd={handleAddBlock} />
        </div>
      </div>

      {/* Properties Panel */}
      {selectedBlock && (
        <div className="w-80 border-l bg-muted/30 p-4 overflow-auto">
          <BlockEditor
            block={selectedBlock}
            onChange={(content) => handleBlockChange(selectedBlock.id, content)}
            onClose={() => setSelectedBlockId(null)}
          />
        </div>
      )}
    </div>
  );
};

// Sortable Block Component
interface SortableBlockProps {
  block: Block;
  isSelected: boolean;
  onSelect: () => void;
  onDelete: () => void;
  onToggleVisibility: () => void;
}

const SortableBlock: React.FC<SortableBlockProps> = ({
  block,
  isSelected,
  onSelect,
  onDelete,
  onToggleVisibility,
}) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: block.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`relative group ${isSelected ? 'ring-2 ring-primary' : ''}`}
      onClick={onSelect}
    >
      {/* Drag Handle */}
      <div
        {...attributes}
        {...listeners}
        className="absolute -left-8 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 cursor-grab"
      >
        <GripVertical className="w-5 h-5 text-muted-foreground" />
      </div>

      {/* Block Actions */}
      <div className="absolute -right-2 top-2 opacity-0 group-hover:opacity-100 flex gap-1">
        <button
          onClick={(e) => {
            e.stopPropagation();
            onToggleVisibility();
          }}
          className="p-1 rounded bg-background shadow"
        >
          {block.is_visible ? (
            <Eye className="w-4 h-4" />
          ) : (
            <EyeOff className="w-4 h-4 text-muted-foreground" />
          )}
        </button>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="p-1 rounded bg-background shadow text-destructive"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      {/* Block Content */}
      <div className={!block.is_visible ? 'opacity-50' : ''}>
        <BlockRenderer block={block} />
      </div>
    </div>
  );
};

// Default content for new blocks
function getDefaultContent(type: string): any {
  switch (type) {
    case 'link':
      return { url: '', title: 'New Link', style: 'default' };
    case 'header':
      return { text: 'Header', level: 2, align: 'center' };
    case 'text':
      return { text: 'Add your text here...', align: 'center' };
    case 'image':
      return { url: '', alt: '', width: 'full' };
    case 'divider':
      return { style: 'line' };
    default:
      return {};
  }
}
```

### Block Renderer

```typescript
// src/components/biopages/BlockRenderer.tsx
import React from 'react';
import { Block } from '@/types/biopage';

interface BlockRendererProps {
  block: Block;
  isPreview?: boolean;
}

export const BlockRenderer: React.FC<BlockRendererProps> = ({ block, isPreview }) => {
  switch (block.type) {
    case 'link':
      return <LinkBlock content={block.content} isPreview={isPreview} />;
    case 'header':
      return <HeaderBlock content={block.content} />;
    case 'text':
      return <TextBlock content={block.content} />;
    case 'image':
      return <ImageBlock content={block.content} />;
    case 'video':
      return <VideoBlock content={block.content} />;
    case 'divider':
      return <DividerBlock content={block.content} />;
    case 'spotify':
      return <SpotifyBlock content={block.content} />;
    case 'youtube':
      return <YouTubeBlock content={block.content} />;
    default:
      return <div>Unknown block type: {block.type}</div>;
  }
};

// Link Block
interface LinkBlockContent {
  url: string;
  title: string;
  description?: string;
  thumbnail?: string;
  style: 'default' | 'featured' | 'outline';
}

const LinkBlock: React.FC<{ content: LinkBlockContent; isPreview?: boolean }> = ({
  content,
  isPreview,
}) => {
  const Component = isPreview ? 'div' : 'a';

  return (
    <Component
      href={isPreview ? undefined : content.url}
      target="_blank"
      rel="noopener noreferrer"
      className={`
        bio-link block w-full p-4 rounded-lg transition-all duration-200
        ${content.style === 'featured' ? 'bg-primary text-primary-foreground' : ''}
        ${content.style === 'outline' ? 'border-2 border-current bg-transparent' : ''}
        ${content.style === 'default' ? 'bg-secondary' : ''}
      `}
    >
      <div className="flex items-center gap-3">
        {content.thumbnail && (
          <img
            src={content.thumbnail}
            alt=""
            className="w-12 h-12 rounded object-cover"
          />
        )}
        <div className="flex-1 text-left">
          <div className="font-medium">{content.title || 'Untitled Link'}</div>
          {content.description && (
            <div className="text-sm opacity-70">{content.description}</div>
          )}
        </div>
      </div>
    </Component>
  );
};

// Header Block
interface HeaderBlockContent {
  text: string;
  level: 1 | 2 | 3;
  align: 'left' | 'center' | 'right';
}

const HeaderBlock: React.FC<{ content: HeaderBlockContent }> = ({ content }) => {
  const Tag = `h${content.level}` as keyof JSX.IntrinsicElements;
  const sizes = { 1: 'text-3xl', 2: 'text-2xl', 3: 'text-xl' };

  return (
    <Tag
      className={`${sizes[content.level]} font-bold text-${content.align}`}
      style={{ textAlign: content.align }}
    >
      {content.text}
    </Tag>
  );
};

// Text Block
interface TextBlockContent {
  text: string;
  align: 'left' | 'center' | 'right';
}

const TextBlock: React.FC<{ content: TextBlockContent }> = ({ content }) => {
  return (
    <p className="whitespace-pre-wrap" style={{ textAlign: content.align }}>
      {content.text}
    </p>
  );
};

// Image Block
interface ImageBlockContent {
  url: string;
  alt: string;
  link_url?: string;
  width: 'full' | 'medium' | 'small';
}

const ImageBlock: React.FC<{ content: ImageBlockContent }> = ({ content }) => {
  const widthClasses = {
    full: 'w-full',
    medium: 'w-3/4',
    small: 'w-1/2',
  };

  const img = (
    <img
      src={content.url}
      alt={content.alt}
      className={`${widthClasses[content.width]} rounded-lg mx-auto`}
    />
  );

  if (content.link_url) {
    return (
      <a href={content.link_url} target="_blank" rel="noopener noreferrer">
        {img}
      </a>
    );
  }

  return img;
};

// Video Block
interface VideoBlockContent {
  url: string;
  provider: string;
  video_id: string;
}

const VideoBlock: React.FC<{ content: VideoBlockContent }> = ({ content }) => {
  const getEmbedUrl = () => {
    switch (content.provider) {
      case 'youtube':
        return `https://www.youtube.com/embed/${content.video_id}`;
      case 'vimeo':
        return `https://player.vimeo.com/video/${content.video_id}`;
      default:
        return content.url;
    }
  };

  return (
    <div className="aspect-video rounded-lg overflow-hidden">
      <iframe
        src={getEmbedUrl()}
        className="w-full h-full"
        allowFullScreen
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
      />
    </div>
  );
};

// Divider Block
const DividerBlock: React.FC<{ content: { style: string } }> = ({ content }) => {
  return <hr className="my-4 border-current opacity-20" />;
};

// Spotify Block
interface SpotifyBlockContent {
  embed_url: string;
  type: 'track' | 'album' | 'playlist' | 'artist';
}

const SpotifyBlock: React.FC<{ content: SpotifyBlockContent }> = ({ content }) => {
  const height = content.type === 'track' ? 152 : 352;

  return (
    <iframe
      src={content.embed_url}
      width="100%"
      height={height}
      frameBorder="0"
      allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
      loading="lazy"
      className="rounded-xl"
    />
  );
};

// YouTube Block
interface YouTubeBlockContent {
  video_id: string;
}

const YouTubeBlock: React.FC<{ content: YouTubeBlockContent }> = ({ content }) => {
  return (
    <div className="aspect-video rounded-lg overflow-hidden">
      <iframe
        src={`https://www.youtube.com/embed/${content.video_id}`}
        className="w-full h-full"
        allowFullScreen
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
      />
    </div>
  );
};
```

---

## Custom CSS

```go
// internal/biopages/css.go
package biopages

import (
	"regexp"
	"strings"
)

// CSSValidator validates and sanitizes custom CSS
type CSSValidator struct {
	allowedProperties []string
	blockedPatterns   []*regexp.Regexp
}

// NewCSSValidator creates a new CSS validator
func NewCSSValidator() *CSSValidator {
	return &CSSValidator{
		allowedProperties: []string{
			"color", "background", "background-color", "background-image",
			"font-family", "font-size", "font-weight", "font-style",
			"text-align", "text-decoration", "line-height", "letter-spacing",
			"margin", "margin-top", "margin-right", "margin-bottom", "margin-left",
			"padding", "padding-top", "padding-right", "padding-bottom", "padding-left",
			"border", "border-radius", "border-color", "border-width", "border-style",
			"box-shadow", "opacity", "transform", "transition",
			"width", "max-width", "min-width", "height", "max-height", "min-height",
			"display", "flex", "flex-direction", "justify-content", "align-items", "gap",
		},
		blockedPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)javascript:`),
			regexp.MustCompile(`(?i)expression\s*\(`),
			regexp.MustCompile(`(?i)url\s*\(\s*["']?\s*data:`),
			regexp.MustCompile(`(?i)@import`),
			regexp.MustCompile(`(?i)behavior\s*:`),
			regexp.MustCompile(`(?i)-moz-binding`),
		},
	}
}

// Validate checks if CSS is safe
func (v *CSSValidator) Validate(css string) (bool, []string) {
	var errors []string

	// Check for blocked patterns
	for _, pattern := range v.blockedPatterns {
		if pattern.MatchString(css) {
			errors = append(errors, "CSS contains blocked pattern: "+pattern.String())
		}
	}

	return len(errors) == 0, errors
}

// Sanitize removes potentially dangerous CSS
func (v *CSSValidator) Sanitize(css string) string {
	// Remove blocked patterns
	sanitized := css
	for _, pattern := range v.blockedPatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "")
	}

	// Scope all selectors to .bio-page-custom
	sanitized = v.scopeSelectors(sanitized, ".bio-page-custom")

	return sanitized
}

func (v *CSSValidator) scopeSelectors(css, scope string) string {
	// Simple regex-based scoping - in production, use a proper CSS parser
	selectorPattern := regexp.MustCompile(`([^{}]+)\{`)

	return selectorPattern.ReplaceAllStringFunc(css, func(match string) string {
		selector := strings.TrimSuffix(match, "{")
		selector = strings.TrimSpace(selector)

		// Don't scope if already scoped or is @-rule
		if strings.HasPrefix(selector, scope) || strings.HasPrefix(selector, "@") {
			return match
		}

		return scope + " " + selector + " {"
	})
}
```

---

## Media Embedding

```go
// internal/biopages/embeds.go
package biopages

import (
	"fmt"
	"net/url"
	"regexp"
)

// EmbedProvider handles media embedding
type EmbedProvider struct {
	providers map[string]*ProviderConfig
}

// ProviderConfig defines embed provider configuration
type ProviderConfig struct {
	Name          string
	URLPatterns   []*regexp.Regexp
	ExtractID     func(string) string
	GetEmbedURL   func(string) string
	GetEmbedHTML  func(string, EmbedOptions) string
}

// EmbedOptions defines embed customization options
type EmbedOptions struct {
	Width     int
	Height    int
	Autoplay  bool
	Loop      bool
	ShowTitle bool
}

// NewEmbedProvider creates a new embed provider
func NewEmbedProvider() *EmbedProvider {
	ep := &EmbedProvider{
		providers: make(map[string]*ProviderConfig),
	}

	// YouTube
	ep.providers["youtube"] = &ProviderConfig{
		Name: "YouTube",
		URLPatterns: []*regexp.Regexp{
			regexp.MustCompile(`youtube\.com/watch\?v=([a-zA-Z0-9_-]+)`),
			regexp.MustCompile(`youtu\.be/([a-zA-Z0-9_-]+)`),
			regexp.MustCompile(`youtube\.com/embed/([a-zA-Z0-9_-]+)`),
		},
		ExtractID: func(urlStr string) string {
			for _, pattern := range ep.providers["youtube"].URLPatterns {
				matches := pattern.FindStringSubmatch(urlStr)
				if len(matches) > 1 {
					return matches[1]
				}
			}
			return ""
		},
		GetEmbedURL: func(id string) string {
			return fmt.Sprintf("https://www.youtube.com/embed/%s", id)
		},
		GetEmbedHTML: func(id string, opts EmbedOptions) string {
			width := opts.Width
			if width == 0 {
				width = 560
			}
			height := opts.Height
			if height == 0 {
				height = 315
			}
			return fmt.Sprintf(
				`<iframe width="%d" height="%d" src="https://www.youtube.com/embed/%s" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`,
				width, height, id,
			)
		},
	}

	// Spotify
	ep.providers["spotify"] = &ProviderConfig{
		Name: "Spotify",
		URLPatterns: []*regexp.Regexp{
			regexp.MustCompile(`open\.spotify\.com/(track|album|playlist|artist)/([a-zA-Z0-9]+)`),
		},
		ExtractID: func(urlStr string) string {
			pattern := ep.providers["spotify"].URLPatterns[0]
			matches := pattern.FindStringSubmatch(urlStr)
			if len(matches) > 2 {
				return matches[1] + "/" + matches[2]
			}
			return ""
		},
		GetEmbedURL: func(id string) string {
			return fmt.Sprintf("https://open.spotify.com/embed/%s", id)
		},
		GetEmbedHTML: func(id string, opts EmbedOptions) string {
			height := 352
			if strings.HasPrefix(id, "track/") {
				height = 152
			}
			return fmt.Sprintf(
				`<iframe src="https://open.spotify.com/embed/%s" width="100%%" height="%d" frameborder="0" allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture" loading="lazy"></iframe>`,
				id, height,
			)
		},
	}

	// TikTok
	ep.providers["tiktok"] = &ProviderConfig{
		Name: "TikTok",
		URLPatterns: []*regexp.Regexp{
			regexp.MustCompile(`tiktok\.com/@[^/]+/video/(\d+)`),
		},
		ExtractID: func(urlStr string) string {
			pattern := ep.providers["tiktok"].URLPatterns[0]
			matches := pattern.FindStringSubmatch(urlStr)
			if len(matches) > 1 {
				return matches[1]
			}
			return ""
		},
		GetEmbedHTML: func(id string, opts EmbedOptions) string {
			return fmt.Sprintf(
				`<blockquote class="tiktok-embed" data-video-id="%s"><script async src="https://www.tiktok.com/embed.js"></script></blockquote>`,
				id,
			)
		},
	}

	// Twitter/X
	ep.providers["twitter"] = &ProviderConfig{
		Name: "Twitter",
		URLPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?:twitter|x)\.com/[^/]+/status/(\d+)`),
		},
		ExtractID: func(urlStr string) string {
			pattern := ep.providers["twitter"].URLPatterns[0]
			matches := pattern.FindStringSubmatch(urlStr)
			if len(matches) > 1 {
				return matches[1]
			}
			return ""
		},
		GetEmbedHTML: func(id string, opts EmbedOptions) string {
			return fmt.Sprintf(
				`<blockquote class="twitter-tweet"><a href="https://twitter.com/i/status/%s"></a></blockquote><script async src="https://platform.twitter.com/widgets.js"></script>`,
				id,
			)
		},
	}

	return ep
}

// ParseURL identifies provider and extracts embed info
func (ep *EmbedProvider) ParseURL(urlStr string) (*EmbedInfo, error) {
	for name, config := range ep.providers {
		for _, pattern := range config.URLPatterns {
			if pattern.MatchString(urlStr) {
				id := config.ExtractID(urlStr)
				if id == "" {
					continue
				}

				return &EmbedInfo{
					Provider: name,
					ID:       id,
					URL:      urlStr,
					EmbedURL: config.GetEmbedURL(id),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("unsupported URL: %s", urlStr)
}

// GetEmbedHTML generates embed HTML for a URL
func (ep *EmbedProvider) GetEmbedHTML(urlStr string, opts EmbedOptions) (string, error) {
	info, err := ep.ParseURL(urlStr)
	if err != nil {
		return "", err
	}

	config := ep.providers[info.Provider]
	return config.GetEmbedHTML(info.ID, opts), nil
}

// EmbedInfo contains parsed embed information
type EmbedInfo struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
	URL      string `json:"url"`
	EmbedURL string `json:"embed_url"`
}
```

---

## SEO Optimization

```go
// internal/biopages/seo.go
package biopages

import (
	"fmt"
	"html/template"
	"strings"
)

// SEOGenerator generates SEO metadata
type SEOGenerator struct {
	baseURL string
}

// NewSEOGenerator creates a new SEO generator
func NewSEOGenerator(baseURL string) *SEOGenerator {
	return &SEOGenerator{baseURL: baseURL}
}

// GenerateMetaTags generates HTML meta tags for a bio page
func (sg *SEOGenerator) GenerateMetaTags(page *BioPage) template.HTML {
	seo := page.SEOSettings
	var tags strings.Builder

	// Basic meta tags
	title := seo.Title
	if title == "" {
		title = page.Title
	}
	tags.WriteString(fmt.Sprintf(`<title>%s</title>`, template.HTMLEscapeString(title)))
	tags.WriteString("\n")

	description := seo.Description
	if description == "" {
		description = page.Description
	}
	if description != "" {
		tags.WriteString(fmt.Sprintf(`<meta name="description" content="%s">`, template.HTMLEscapeString(description)))
		tags.WriteString("\n")
	}

	// Keywords
	if len(seo.Keywords) > 0 {
		tags.WriteString(fmt.Sprintf(`<meta name="keywords" content="%s">`, template.HTMLEscapeString(strings.Join(seo.Keywords, ", "))))
		tags.WriteString("\n")
	}

	// Canonical URL
	canonicalURL := seo.CanonicalURL
	if canonicalURL == "" {
		canonicalURL = fmt.Sprintf("%s/%s", sg.baseURL, page.Slug)
	}
	tags.WriteString(fmt.Sprintf(`<link rel="canonical" href="%s">`, canonicalURL))
	tags.WriteString("\n")

	// Robots
	if seo.NoIndex {
		tags.WriteString(`<meta name="robots" content="noindex, nofollow">`)
		tags.WriteString("\n")
	}

	// Open Graph tags
	tags.WriteString(fmt.Sprintf(`<meta property="og:title" content="%s">`, template.HTMLEscapeString(title)))
	tags.WriteString("\n")
	tags.WriteString(fmt.Sprintf(`<meta property="og:description" content="%s">`, template.HTMLEscapeString(description)))
	tags.WriteString("\n")
	tags.WriteString(fmt.Sprintf(`<meta property="og:url" content="%s">`, canonicalURL))
	tags.WriteString("\n")
	tags.WriteString(`<meta property="og:type" content="website">`)
	tags.WriteString("\n")

	ogImage := seo.OGImage
	if ogImage == "" && page.AvatarURL != "" {
		ogImage = page.AvatarURL
	}
	if ogImage != "" {
		tags.WriteString(fmt.Sprintf(`<meta property="og:image" content="%s">`, ogImage))
		tags.WriteString("\n")
	}

	// Twitter Card
	twitterCard := seo.TwitterCard
	if twitterCard == "" {
		twitterCard = "summary"
	}
	tags.WriteString(fmt.Sprintf(`<meta name="twitter:card" content="%s">`, twitterCard))
	tags.WriteString("\n")
	tags.WriteString(fmt.Sprintf(`<meta name="twitter:title" content="%s">`, template.HTMLEscapeString(title)))
	tags.WriteString("\n")
	tags.WriteString(fmt.Sprintf(`<meta name="twitter:description" content="%s">`, template.HTMLEscapeString(description)))
	tags.WriteString("\n")
	if ogImage != "" {
		tags.WriteString(fmt.Sprintf(`<meta name="twitter:image" content="%s">`, ogImage))
		tags.WriteString("\n")
	}

	return template.HTML(tags.String())
}

// GenerateStructuredData generates JSON-LD structured data
func (sg *SEOGenerator) GenerateStructuredData(page *BioPage) template.JS {
	canonicalURL := page.SEOSettings.CanonicalURL
	if canonicalURL == "" {
		canonicalURL = fmt.Sprintf("%s/%s", sg.baseURL, page.Slug)
	}

	// Person/Organization schema
	data := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "ProfilePage",
		"name":     page.Title,
		"url":      canonicalURL,
	}

	if page.Description != "" {
		data["description"] = page.Description
	}

	if page.AvatarURL != "" {
		data["image"] = page.AvatarURL
	}

	// Add social links as sameAs
	var sameAs []string
	for _, social := range page.SocialLinks {
		sameAs = append(sameAs, social.URL)
	}
	if len(sameAs) > 0 {
		data["mainEntity"] = map[string]interface{}{
			"@type":  "Person",
			"name":   page.Title,
			"sameAs": sameAs,
		}
	}

	jsonData, _ := json.Marshal(data)
	return template.JS(jsonData)
}
```

---

## API Endpoints

```go
// internal/api/handlers/biopages.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/biopages"
)

// BioPageHandler handles bio page API requests
type BioPageHandler struct {
	service *biopages.Service
}

// RegisterRoutes registers bio page routes
func (h *BioPageHandler) RegisterRoutes(app *fiber.App) {
	pages := app.Group("/api/v1/biopages")

	pages.Get("/", h.ListPages)
	pages.Post("/", h.CreatePage)
	pages.Get("/:id", h.GetPage)
	pages.Put("/:id", h.UpdatePage)
	pages.Delete("/:id", h.DeletePage)
	pages.Post("/:id/publish", h.PublishPage)
	pages.Post("/:id/unpublish", h.UnpublishPage)
	pages.Get("/:id/analytics", h.GetAnalytics)

	// Blocks
	pages.Post("/:id/blocks", h.AddBlock)
	pages.Put("/:id/blocks/:blockId", h.UpdateBlock)
	pages.Delete("/:id/blocks/:blockId", h.DeleteBlock)
	pages.Post("/:id/blocks/reorder", h.ReorderBlocks)

	// Themes
	app.Get("/api/v1/themes", h.ListThemes)
	app.Get("/api/v1/themes/:id", h.GetTheme)
	app.Post("/api/v1/themes", h.CreateCustomTheme)

	// Public page rendering
	app.Get("/bio/:slug", h.RenderPublicPage)
}

// CreatePage creates a new bio page
func (h *BioPageHandler) CreatePage(c *fiber.Ctx) error {
	var req CreateBioPageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)

	page, err := h.service.Create(c.Context(), workspaceID, &biopages.CreatePageInput{
		Title:       req.Title,
		Description: req.Description,
		Slug:        req.Slug,
		ThemeID:     req.ThemeID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(page)
}

// UpdatePage updates a bio page
func (h *BioPageHandler) UpdatePage(c *fiber.Ctx) error {
	pageID := c.Params("id")

	var req UpdateBioPageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	page, err := h.service.Update(c.Context(), pageID, &biopages.UpdatePageInput{
		Title:       req.Title,
		Description: req.Description,
		Slug:        req.Slug,
		ThemeID:     req.ThemeID,
		CustomCSS:   req.CustomCSS,
		AvatarURL:   req.AvatarURL,
		SEOSettings: req.SEOSettings,
		SocialLinks: req.SocialLinks,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(page)
}

// ReorderBlocks reorders blocks on a page
func (h *BioPageHandler) ReorderBlocks(c *fiber.Ctx) error {
	pageID := c.Params("id")

	var req ReorderBlocksRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.service.ReorderBlocks(c.Context(), pageID, req.BlockIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true})
}

// RenderPublicPage renders the public bio page
func (h *BioPageHandler) RenderPublicPage(c *fiber.Ctx) error {
	slug := c.Params("slug")

	page, err := h.service.GetBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Page not found")
	}

	if !page.IsPublished {
		return c.Status(fiber.StatusNotFound).SendString("Page not found")
	}

	// Track page view
	go h.service.TrackPageView(context.Background(), page.ID, c.IP(), c.Get("User-Agent"))

	// Render HTML
	html, err := h.service.RenderPage(c.Context(), page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error rendering page")
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// Request types
type CreateBioPageRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Slug        string `json:"slug" validate:"required,slug"`
	ThemeID     string `json:"theme_id"`
}

type UpdateBioPageRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Slug        string               `json:"slug"`
	ThemeID     string               `json:"theme_id"`
	CustomCSS   string               `json:"custom_css"`
	AvatarURL   string               `json:"avatar_url"`
	SEOSettings *biopages.SEOSettings `json:"seo_settings"`
	SocialLinks []biopages.SocialLink `json:"social_links"`
}

type ReorderBlocksRequest struct {
	BlockIDs []string `json:"block_ids" validate:"required"`
}
```

---

## React Components

### Bio Page Preview

```typescript
// src/components/biopages/BioPagePreview.tsx
import React from 'react';
import { BioPage, Block } from '@/types/biopage';
import { BlockRenderer } from './BlockRenderer';

interface BioPagePreviewProps {
  page: BioPage;
  theme?: Theme;
}

export const BioPagePreview: React.FC<BioPagePreviewProps> = ({ page, theme }) => {
  const visibleBlocks = page.blocks
    .filter((block) => block.is_visible)
    .sort((a, b) => a.order - b.order);

  return (
    <div
      className="bio-page min-h-screen py-8 px-4"
      style={getThemeStyles(theme)}
    >
      <div className="bio-content max-w-md mx-auto">
        {/* Avatar */}
        {page.avatar_url && (
          <div className="flex justify-center mb-6">
            <img
              src={page.avatar_url}
              alt={page.title}
              className="bio-avatar w-24 h-24 rounded-full object-cover"
            />
          </div>
        )}

        {/* Title */}
        <h1 className="text-2xl font-bold text-center mb-2">{page.title}</h1>

        {/* Description */}
        {page.description && (
          <p className="text-center mb-6 opacity-80">{page.description}</p>
        )}

        {/* Social Links */}
        {page.social_links?.length > 0 && (
          <div className="flex justify-center gap-4 mb-8">
            {page.social_links.map((social, index) => (
              <a
                key={index}
                href={social.url}
                target="_blank"
                rel="noopener noreferrer"
                className="social-icon hover:opacity-80 transition-opacity"
              >
                <SocialIcon platform={social.platform} />
              </a>
            ))}
          </div>
        )}

        {/* Blocks */}
        <div className="space-y-4">
          {visibleBlocks.map((block) => (
            <BlockRenderer key={block.id} block={block} />
          ))}
        </div>

        {/* Footer */}
        <div className="mt-12 text-center text-sm opacity-50">
          <a href="https://linkrift.io" target="_blank" rel="noopener noreferrer">
            Powered by Linkrift
          </a>
        </div>
      </div>

      {/* Custom CSS */}
      {page.custom_css && (
        <style dangerouslySetInnerHTML={{ __html: page.custom_css }} />
      )}
    </div>
  );
};

function getThemeStyles(theme?: Theme): React.CSSProperties {
  if (!theme) return {};

  const styles: React.CSSProperties = {
    fontFamily: theme.styles.font_family,
    color: theme.styles.text_color,
  };

  switch (theme.styles.background_type) {
    case 'solid':
      styles.backgroundColor = theme.styles.background_color;
      break;
    case 'gradient':
      if (theme.styles.background_gradient) {
        const g = theme.styles.background_gradient;
        const colorStops = g.colors
          .map((color, i) => `${color} ${g.positions[i] || 0}%`)
          .join(', ');
        styles.background =
          g.type === 'radial'
            ? `radial-gradient(circle, ${colorStops})`
            : `linear-gradient(${g.angle}deg, ${colorStops})`;
      }
      break;
    case 'image':
      styles.backgroundImage = `url(${theme.styles.background_image})`;
      styles.backgroundSize = 'cover';
      styles.backgroundPosition = 'center';
      break;
  }

  return styles;
}

// Social Icon Component
const SocialIcon: React.FC<{ platform: string }> = ({ platform }) => {
  const icons: Record<string, JSX.Element> = {
    twitter: <TwitterIcon />,
    instagram: <InstagramIcon />,
    facebook: <FacebookIcon />,
    linkedin: <LinkedInIcon />,
    youtube: <YouTubeIcon />,
    tiktok: <TikTokIcon />,
    github: <GitHubIcon />,
    email: <EmailIcon />,
  };

  return icons[platform.toLowerCase()] || <LinkIcon />;
};
```

### Theme Selector

```typescript
// src/components/biopages/ThemeSelector.tsx
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { biopagesApi, Theme } from '@/api/biopages';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Check, Crown } from 'lucide-react';

interface ThemeSelectorProps {
  selectedThemeId: string;
  onSelect: (themeId: string) => void;
  isPremiumUser?: boolean;
}

export const ThemeSelector: React.FC<ThemeSelectorProps> = ({
  selectedThemeId,
  onSelect,
  isPremiumUser = false,
}) => {
  const { data: themes, isLoading } = useQuery({
    queryKey: ['themes'],
    queryFn: biopagesApi.listThemes,
  });

  if (isLoading) {
    return <div>Loading themes...</div>;
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
      {themes?.map((theme) => {
        const isSelected = theme.id === selectedThemeId;
        const isLocked = theme.is_premium && !isPremiumUser;

        return (
          <Card
            key={theme.id}
            className={`cursor-pointer transition-all ${
              isSelected ? 'ring-2 ring-primary' : ''
            } ${isLocked ? 'opacity-60' : ''}`}
            onClick={() => !isLocked && onSelect(theme.id)}
          >
            <CardContent className="p-4">
              {/* Theme Preview */}
              <div
                className="aspect-[9/16] rounded-lg mb-3 relative overflow-hidden"
                style={getPreviewStyles(theme)}
              >
                {/* Mock content */}
                <div className="absolute inset-0 flex flex-col items-center justify-start p-4 pt-8">
                  <div
                    className="w-12 h-12 rounded-full mb-3"
                    style={{ backgroundColor: theme.styles.button_color + '40' }}
                  />
                  <div
                    className="w-24 h-3 rounded mb-2"
                    style={{ backgroundColor: theme.styles.title_color }}
                  />
                  <div
                    className="w-16 h-2 rounded mb-6"
                    style={{ backgroundColor: theme.styles.text_color + '60' }}
                  />
                  {[1, 2, 3].map((i) => (
                    <div
                      key={i}
                      className="w-full h-8 rounded mb-2"
                      style={{
                        backgroundColor: theme.styles.button_color,
                        borderRadius: theme.styles.button_border_radius,
                      }}
                    />
                  ))}
                </div>

                {/* Selected indicator */}
                {isSelected && (
                  <div className="absolute top-2 right-2 w-6 h-6 bg-primary rounded-full flex items-center justify-center">
                    <Check className="w-4 h-4 text-primary-foreground" />
                  </div>
                )}

                {/* Premium indicator */}
                {theme.is_premium && (
                  <div className="absolute top-2 left-2">
                    <Badge variant="secondary" className="gap-1">
                      <Crown className="w-3 h-3" />
                      Pro
                    </Badge>
                  </div>
                )}
              </div>

              {/* Theme name */}
              <p className="font-medium text-sm text-center">{theme.name}</p>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
};

function getPreviewStyles(theme: Theme): React.CSSProperties {
  const styles: React.CSSProperties = {};

  switch (theme.styles.background_type) {
    case 'solid':
      styles.backgroundColor = theme.styles.background_color;
      break;
    case 'gradient':
      if (theme.styles.background_gradient) {
        const g = theme.styles.background_gradient;
        const colorStops = g.colors
          .map((color, i) => `${color} ${g.positions[i] || 0}%`)
          .join(', ');
        styles.background = `linear-gradient(${g.angle}deg, ${colorStops})`;
      }
      break;
  }

  return styles;
}
```
