# How to Submit This Project

A complete step-by-step guide to submitting your completed assignment.

---

## What You'll Need

Before starting, make sure you have:
- ✅ A GitHub account (free at github.com)
- ✅ Git installed on your computer
- ✅ The completed project code

---

## Step 1: Create a Private Repository on GitHub

### 1.1 Open GitHub
Go to **https://github.com** in your web browser and log in to your account.

### 1.2 Click "New Repository"
- Look for a **green button** that says **"New"** or **"+"** in the top-right corner
- Click it and select **"New repository"**

### 1.3 Fill in the Details

| Field | What to Enter |
|-------|---------------|
| **Repository name** | `timetravel-assignment` (or any name you prefer) |
| **Description** | `Rainbow Backend Take-Home Assignment` (optional) |
| **Visibility** | 🔴 **IMPORTANT: Select "Private"** |
| **Initialize** | Leave everything unchecked |

### 1.4 Click "Create repository"
You'll see a page with setup instructions. **Keep this page open** - you'll need the URL shown.

---

## Step 2: Open Your Terminal

### On Mac:
1. Press **Command + Space** to open Spotlight
2. Type **"Terminal"**
3. Press **Enter**

### Navigate to Your Project:
Type this command and press Enter:
```bash
cd /Users/jayanthvishalreddy/Downloads/timetravel
```

---

## Step 3: Connect Your Code to GitHub

### 3.1 Check if Git is Already Set Up
Type this and press Enter:
```bash
git remote -v
```

**If you see "origin" pointing to rainbowmga**, run:
```bash
git remote rename origin upstream
```

### 3.2 Add Your Private Repository
Replace `YOUR_USERNAME` with your GitHub username and `YOUR_REPO_NAME` with the repository name you created:

```bash
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
```

**Example:**
```bash
git remote add origin https://github.com/johndoe/timetravel-assignment.git
```

---

## Step 4: Save Your Changes (Commit)

### 4.1 Tell Git Who You Are (First Time Only)
```bash
git config user.email "your-email@example.com"
git config user.name "Your Name"
```

### 4.2 Add All Your Files
```bash
git add .
```

### 4.3 Create a Commit
```bash
git commit -m "Complete Rainbow Backend Assignment - SQLite persistence and time-travel versioning"
```

**What this means:**
- `git commit` = Save a snapshot of your code
- `-m "..."` = A message describing what you did

---

## Step 5: Upload to GitHub (Push)

```bash
git push -u origin master
```

**You might be asked to log in:**
- Enter your GitHub username
- For password, use a **Personal Access Token** (not your regular password)

### How to Create a Personal Access Token:
1. Go to **https://github.com/settings/tokens**
2. Click **"Generate new token (classic)"**
3. Give it a name like "Project Upload"
4. Check the box for **"repo"** (full control)
5. Click **"Generate token"**
6. **Copy the token** and paste it when asked for password

---

## Step 6: Verify Your Upload

### 6.1 Check on GitHub
1. Go to **https://github.com/YOUR_USERNAME/YOUR_REPO_NAME**
2. You should see all your code files listed

### 6.2 Make Sure It's Private
- Look for a **"Private"** label next to the repository name
- If it says "Public", click **Settings → Danger Zone → Change visibility**

---

## Step 7: Invite Rainbow to Your Repository

### 7.1 Go to Repository Settings
1. On your repository page, click **"Settings"** (gear icon)
2. Click **"Collaborators"** in the left sidebar

### 7.2 Add Collaborators
1. Click **"Add people"**
2. Type the email or GitHub username provided by Rainbow
3. Click **"Add"**

---

## Step 8: Send the Email

Write an email to Rainbow with:

**Subject:** Backend Take-Home Assignment Submission

**Body:**
```
Hi,

I have completed the Backend Take-Home Assignment. 

Repository Link: https://github.com/YOUR_USERNAME/YOUR_REPO_NAME

What I built:
- Objective 1: Replaced in-memory storage with SQLite persistence
- Objective 2: Implemented time-travel versioning with v2 API

The repository is private. I've added [their username] as a collaborator.

Please let me know if you need any additional information.

Thank you,
[Your Name]
```

---

## Quick Reference: All Commands

```bash
# Navigate to project
cd /Users/jayanthvishalreddy/Downloads/timetravel

# Set up git (if needed)
git remote rename origin upstream
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git

# Save and upload
git add .
git commit -m "Complete Rainbow Backend Assignment"
git push -u origin master
```

---

## Troubleshooting

### "Permission denied" error
→ Make sure you're using a Personal Access Token, not your password

### "Repository not found" error
→ Double-check the URL matches exactly what GitHub shows

### "Nothing to commit" message
→ Your changes are already saved! Just run `git push`

### "Branch 'master' not found" error
Try:
```bash
git push -u origin main
```
(Some repos use "main" instead of "master")

---

## Checklist Before Submitting

- [ ] All code files are in the repository
- [ ] Repository is set to **Private**
- [ ] You can see files when you visit the GitHub URL
- [ ] Rainbow team has been added as collaborators
- [ ] Email has been sent with the repository link

---

**You're done! 🎉**
