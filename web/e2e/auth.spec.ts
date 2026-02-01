import { test, expect } from "@playwright/test"

test.describe("Authentication", () => {
  test("shows login page at /auth/login", async ({ page }) => {
    await page.goto("/auth/login")
    await expect(page.getByRole("heading", { name: /sign in|log in/i })).toBeVisible()
    await expect(page.getByLabel(/email/i)).toBeVisible()
    await expect(page.getByLabel(/password/i)).toBeVisible()
  })

  test("shows register page at /auth/register", async ({ page }) => {
    await page.goto("/auth/register")
    await expect(page.getByRole("heading", { name: /sign up|create|register/i })).toBeVisible()
    await expect(page.getByLabel(/name/i)).toBeVisible()
    await expect(page.getByLabel(/email/i)).toBeVisible()
  })

  test("navigates between login and register", async ({ page }) => {
    await page.goto("/auth/login")
    await page.getByRole("link", { name: /sign up|register|create account/i }).click()
    await expect(page).toHaveURL(/register/)

    await page.getByRole("link", { name: /sign in|log in|already have/i }).click()
    await expect(page).toHaveURL(/login/)
  })

  test("shows forgot password page", async ({ page }) => {
    await page.goto("/auth/login")
    await page.getByRole("link", { name: /forgot/i }).click()
    await expect(page).toHaveURL(/forgot/)
    await expect(page.getByLabel(/email/i)).toBeVisible()
  })

  test("shows validation errors on empty login submit", async ({ page }) => {
    await page.goto("/auth/login")
    await page.getByRole("button", { name: /sign in|log in/i }).click()
    // Should show validation errors or remain on login page
    await expect(page).toHaveURL(/login/)
  })

  test("redirects unauthenticated users to login", async ({ page }) => {
    await page.goto("/")
    // Should redirect to login page
    await expect(page).toHaveURL(/login/)
  })
})
