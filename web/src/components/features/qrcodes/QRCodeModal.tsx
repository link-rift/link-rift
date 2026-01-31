import { useState, useEffect, useCallback } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { FeatureGate } from "@/components/ui/FeatureGate"
import {
  useQRCodeForLink,
  useCreateQRCode,
  useDownloadQRCode,
  useStyleTemplates,
} from "@/hooks/useQRCodes"
import type { Link } from "@/types/link"
import type { CreateQRCodeRequest } from "@/types/qrcode"

interface QRCodeModalProps {
  link: Link
  open: boolean
  onClose: () => void
}

const DEFAULT_OPTIONS: CreateQRCodeRequest = {
  qr_type: "dynamic",
  error_correction: "M",
  foreground_color: "#000000",
  background_color: "#FFFFFF",
  dot_style: "square",
  corner_style: "square",
  size: 512,
  margin: 4,
}

export default function QRCodeModal({ link, open, onClose }: QRCodeModalProps) {
  const [options, setOptions] = useState<CreateQRCodeRequest>(DEFAULT_OPTIONS)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)

  const { data: existingQR } = useQRCodeForLink(link.id)
  const { data: templates } = useStyleTemplates()
  const createQR = useCreateQRCode()
  const downloadQR = useDownloadQRCode()

  // Load existing QR settings if available
  useEffect(() => {
    if (existingQR) {
      setOptions({
        qr_type: existingQR.qr_type,
        error_correction: existingQR.error_correction,
        foreground_color: existingQR.foreground_color,
        background_color: existingQR.background_color,
        dot_style: existingQR.dot_style,
        corner_style: existingQR.corner_style,
        size: existingQR.size,
        margin: existingQR.margin,
      })
      if (existingQR.png_url) {
        setPreviewUrl(existingQR.png_url)
      }
    }
  }, [existingQR])

  // Generate a simple preview using canvas
  const generatePreview = useCallback(() => {
    const canvas = document.createElement("canvas")
    const size = 200
    canvas.width = size
    canvas.height = size
    const ctx = canvas.getContext("2d")
    if (!ctx) return

    const fg = options.foreground_color || "#000000"
    const bg = options.background_color || "#FFFFFF"

    // Background
    ctx.fillStyle = bg
    ctx.fillRect(0, 0, size, size)

    // Simple QR placeholder pattern
    const moduleSize = 6
    const margin = (options.margin || 4) * 2
    const offset = margin

    ctx.fillStyle = fg

    // Draw finder patterns (top-left, top-right, bottom-left)
    const drawFinder = (x: number, y: number) => {
      // Outer border
      ctx.fillRect(x, y, moduleSize * 7, moduleSize * 7)
      // Inner white
      ctx.fillStyle = bg
      ctx.fillRect(
        x + moduleSize,
        y + moduleSize,
        moduleSize * 5,
        moduleSize * 5
      )
      // Inner dark
      ctx.fillStyle = fg
      ctx.fillRect(
        x + moduleSize * 2,
        y + moduleSize * 2,
        moduleSize * 3,
        moduleSize * 3
      )
    }

    drawFinder(offset, offset)
    drawFinder(size - offset - moduleSize * 7, offset)
    drawFinder(offset, size - offset - moduleSize * 7)

    // Draw some data modules (pseudo-random pattern based on link short_code)
    const seed = link.short_code
      .split("")
      .reduce((a, c) => a + c.charCodeAt(0), 0)
    for (let i = 0; i < 15; i++) {
      for (let j = 0; j < 15; j++) {
        if ((seed + i * 7 + j * 13) % 3 === 0) {
          const mx = offset + moduleSize * 8 + j * moduleSize
          const my = offset + moduleSize * 8 + i * moduleSize
          if (mx + moduleSize < size - offset && my + moduleSize < size - offset) {
            if (options.dot_style === "dots") {
              ctx.beginPath()
              ctx.arc(
                mx + moduleSize / 2,
                my + moduleSize / 2,
                moduleSize / 2,
                0,
                Math.PI * 2
              )
              ctx.fill()
            } else if (options.dot_style === "rounded") {
              const r = moduleSize * 0.3
              ctx.beginPath()
              ctx.moveTo(mx + r, my)
              ctx.lineTo(mx + moduleSize - r, my)
              ctx.quadraticCurveTo(mx + moduleSize, my, mx + moduleSize, my + r)
              ctx.lineTo(mx + moduleSize, my + moduleSize - r)
              ctx.quadraticCurveTo(
                mx + moduleSize,
                my + moduleSize,
                mx + moduleSize - r,
                my + moduleSize
              )
              ctx.lineTo(mx + r, my + moduleSize)
              ctx.quadraticCurveTo(mx, my + moduleSize, mx, my + moduleSize - r)
              ctx.lineTo(mx, my + r)
              ctx.quadraticCurveTo(mx, my, mx + r, my)
              ctx.fill()
            } else {
              ctx.fillRect(mx, my, moduleSize, moduleSize)
            }
          }
        }
      }
    }

    setPreviewUrl(canvas.toDataURL("image/png"))
  }, [options, link.short_code])

  // Generate preview when options change and no existing QR
  useEffect(() => {
    if (!existingQR?.png_url) {
      generatePreview()
    }
  }, [options, existingQR?.png_url, generatePreview])

  function handleCreate() {
    createQR.mutate(
      { linkId: link.id, data: options },
      {
        onSuccess: (qr) => {
          if (qr.png_url) {
            setPreviewUrl(qr.png_url)
          }
        },
      }
    )
  }

  function handleDownload(format: "png" | "svg") {
    downloadQR.mutate(
      { linkId: link.id, format },
      {
        onSuccess: (blob) => {
          const url = URL.createObjectURL(blob)
          const a = document.createElement("a")
          a.href = url
          a.download = `qr-${link.short_code}.${format}`
          document.body.appendChild(a)
          a.click()
          document.body.removeChild(a)
          URL.revokeObjectURL(url)
        },
      }
    )
  }

  function applyTemplate(key: string) {
    if (!templates) return
    const t = templates[key]
    if (!t) return
    setOptions((prev) => ({
      ...prev,
      foreground_color: t.foreground_color,
      background_color: t.background_color,
      dot_style: t.dot_style,
      corner_style: t.corner_style,
      error_correction: t.error_correction,
    }))
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>QR Code</DialogTitle>
          <DialogDescription>
            Generate a QR code for{" "}
            <span className="font-medium text-foreground">
              {link.short_url}
            </span>
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
          {/* Preview */}
          <div className="flex flex-col items-center gap-4">
            <div className="flex h-52 w-52 items-center justify-center rounded-lg border bg-white">
              {previewUrl ? (
                <img
                  src={previewUrl}
                  alt="QR Code preview"
                  className="h-48 w-48 object-contain"
                />
              ) : (
                <div className="text-sm text-muted-foreground">
                  Preview will appear here
                </div>
              )}
            </div>

            <div className="flex gap-2">
              {existingQR ? (
                <>
                  <Button
                    size="sm"
                    onClick={() => handleDownload("png")}
                    disabled={downloadQR.isPending}
                  >
                    {downloadQR.isPending ? "Downloading..." : "PNG"}
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleDownload("svg")}
                    disabled={downloadQR.isPending}
                  >
                    SVG
                  </Button>
                </>
              ) : (
                <Button
                  size="sm"
                  onClick={handleCreate}
                  disabled={createQR.isPending}
                >
                  {createQR.isPending ? "Generating..." : "Generate QR Code"}
                </Button>
              )}
            </div>

            {createQR.isError && (
              <p className="text-sm text-destructive">
                {createQR.error instanceof Error
                  ? createQR.error.message
                  : "Failed to generate QR code"}
              </p>
            )}
          </div>

          {/* Customization */}
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Type</Label>
              <Select
                value={options.qr_type}
                onValueChange={(v) =>
                  setOptions((prev) => ({ ...prev, qr_type: v }))
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="dynamic">Dynamic (trackable)</SelectItem>
                  <SelectItem value="static">Static (direct URL)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <FeatureGate
              feature="qr_customization"
              upgradeVariant="inline"
              fallback={
                <p className="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
                  Upgrade to Pro to customize colors, styles, and error
                  correction.
                </p>
              }
            >
              <Tabs defaultValue="templates">
                <TabsList className="w-full">
                  <TabsTrigger value="templates">Templates</TabsTrigger>
                  <TabsTrigger value="colors">Colors</TabsTrigger>
                  <TabsTrigger value="style">Style</TabsTrigger>
                </TabsList>

                <TabsContent value="templates" className="space-y-2 pt-2">
                  <div className="grid grid-cols-2 gap-2">
                    {templates &&
                      Object.entries(templates).map(([key, t]) => (
                        <button
                          key={key}
                          onClick={() => applyTemplate(key)}
                          className="flex items-center gap-2 rounded-md border p-2 text-left text-sm transition-colors hover:bg-muted/50"
                        >
                          <div
                            className="h-6 w-6 rounded border"
                            style={{
                              backgroundColor: t.foreground_color,
                            }}
                          />
                          <span>{t.name}</span>
                        </button>
                      ))}
                  </div>
                </TabsContent>

                <TabsContent value="colors" className="space-y-3 pt-2">
                  <div className="space-y-1.5">
                    <Label className="text-xs">Foreground</Label>
                    <div className="flex gap-2">
                      <input
                        type="color"
                        value={options.foreground_color}
                        onChange={(e) =>
                          setOptions((prev) => ({
                            ...prev,
                            foreground_color: e.target.value,
                          }))
                        }
                        className="h-9 w-9 cursor-pointer rounded border"
                      />
                      <Input
                        value={options.foreground_color}
                        onChange={(e) =>
                          setOptions((prev) => ({
                            ...prev,
                            foreground_color: e.target.value,
                          }))
                        }
                        className="font-mono text-xs"
                      />
                    </div>
                  </div>
                  <div className="space-y-1.5">
                    <Label className="text-xs">Background</Label>
                    <div className="flex gap-2">
                      <input
                        type="color"
                        value={options.background_color}
                        onChange={(e) =>
                          setOptions((prev) => ({
                            ...prev,
                            background_color: e.target.value,
                          }))
                        }
                        className="h-9 w-9 cursor-pointer rounded border"
                      />
                      <Input
                        value={options.background_color}
                        onChange={(e) =>
                          setOptions((prev) => ({
                            ...prev,
                            background_color: e.target.value,
                          }))
                        }
                        className="font-mono text-xs"
                      />
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="style" className="space-y-3 pt-2">
                  <div className="space-y-1.5">
                    <Label className="text-xs">
                      Size: {options.size}px
                    </Label>
                    <input
                      type="range"
                      min={128}
                      max={1024}
                      step={64}
                      value={options.size}
                      onChange={(e) =>
                        setOptions((prev) => ({
                          ...prev,
                          size: Number(e.target.value),
                        }))
                      }
                      className="w-full"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <Label className="text-xs">Dot Style</Label>
                    <Select
                      value={options.dot_style}
                      onValueChange={(v) =>
                        setOptions((prev) => ({ ...prev, dot_style: v }))
                      }
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="square">Square</SelectItem>
                        <SelectItem value="rounded">Rounded</SelectItem>
                        <SelectItem value="dots">Dots</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1.5">
                    <Label className="text-xs">Corner Style</Label>
                    <Select
                      value={options.corner_style}
                      onValueChange={(v) =>
                        setOptions((prev) => ({
                          ...prev,
                          corner_style: v,
                        }))
                      }
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="square">Square</SelectItem>
                        <SelectItem value="rounded">Rounded</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1.5">
                    <Label className="text-xs">Error Correction</Label>
                    <Select
                      value={options.error_correction}
                      onValueChange={(v) =>
                        setOptions((prev) => ({
                          ...prev,
                          error_correction: v,
                        }))
                      }
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="L">Low (7%)</SelectItem>
                        <SelectItem value="M">Medium (15%)</SelectItem>
                        <SelectItem value="Q">Quartile (25%)</SelectItem>
                        <SelectItem value="H">High (30%)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1.5">
                    <Label className="text-xs">
                      Margin: {options.margin}
                    </Label>
                    <input
                      type="range"
                      min={0}
                      max={10}
                      step={1}
                      value={options.margin}
                      onChange={(e) =>
                        setOptions((prev) => ({
                          ...prev,
                          margin: Number(e.target.value),
                        }))
                      }
                      className="w-full"
                    />
                  </div>
                </TabsContent>
              </Tabs>
            </FeatureGate>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
