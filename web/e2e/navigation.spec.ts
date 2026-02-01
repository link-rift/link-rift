import { test, expect } from "@playwright/test"

test.describe("Navigation", () => {
  test("login page renders without errors", async ({ page }) => {
    await page.goto("/auth/login")
    // No console errors
    const errors: string[] = []
    page.on("pageerror", (err) => errors.push(err.message))
    await page.waitForTimeout(1000)
    expect(errors).toHaveLength(0)
  })

  test("register page renders without errors", async ({ page }) => {
    await page.goto("/auth/register")
    const errors: string[] = []
    page.on("pageerror", (err) => errors.push(err.message))
    await page.waitForTimeout(1000)
    expect(errors).toHaveLength(0)
  })

  test("forgot password page renders without errors", async ({ page }) => {
    await page.goto("/auth/forgot-password")
    const errors: string[] = []
    page.on("pageerror", (err) => errors.push(err.message))
    await page.waitForTimeout(1000)
    expect(errors).toHaveLength(0)
  })

  test("app has correct title", async ({ page }) => {
    await page.goto("/auth/login")
    await expect(page).toHaveTitle(/linkrift/i)
  })
})
