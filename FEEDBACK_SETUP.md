# Feedback Feature Setup

The feedback feature allows users to submit feature ideas, bug reports, and improvements directly from the app. Submissions are automatically created as GitHub Issues in your repository.

## Features

- 💡 **Feature Idea Button**: Located next to the nuke buttons in the bottom-left corner
- 📝 **Three Categories**: Bug reports, feature requests, and improvements
- 🏷️ **Auto-labeling**: Issues are automatically tagged with appropriate labels
- ✅ **Validation**: Input validation for title (5-100 chars) and description (10-1000 chars)
- 🔒 **Rate Limiting**: Built-in protection against spam

## Setup Instructions

### 1. Create a GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name like "BreakoutGlobe Feedback"
4. Select the following scopes:
   - `repo` (Full control of private repositories)
   - Or just `public_repo` if your repository is public
5. Click "Generate token"
6. **Copy the token immediately** (you won't be able to see it again)

### 2. Configure Environment Variables

Add these environment variables to your backend:

```bash
# GitHub Integration for Feedback
GITHUB_TOKEN=ghp_your_token_here
GITHUB_REPO_OWNER=your-github-username
GITHUB_REPO_NAME=your-repo-name
```

#### Local Development (.env file)
```bash
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
GITHUB_REPO_OWNER=yourusername
GITHUB_REPO_NAME=breakoutglobe
```

#### Railway Deployment
1. Go to your Railway project
2. Click on your backend service
3. Go to "Variables" tab
4. Add the three environment variables above

### 3. Test the Integration

1. Start your backend server
2. Open the app in your browser
3. Click the "💡 Feature Idea" button
4. Fill out the form and submit
5. Check your GitHub repository's Issues tab

## API Endpoint

**POST** `/api/feedback`

### Request Body
```json
{
  "title": "Add dark mode",
  "description": "It would be great to have a dark mode option",
  "category": "feature"
}
```

### Categories
- `bug` - Bug reports (labeled as "bug")
- `feature` - Feature requests (labeled as "enhancement")
- `improvement` - Improvements (labeled as "enhancement")

All submissions also get the "user-feedback" label.

### Response
```json
{
  "success": true,
  "message": "Feedback submitted successfully"
}
```

## Security Features

- ✅ Input validation (length limits, required fields)
- ✅ Category whitelist
- ✅ Sanitized user input
- ✅ GitHub token stored securely in environment variables
- ✅ Rate limiting to prevent spam

## Troubleshooting

### "GitHub integration is not configured"
- Ensure all three environment variables are set
- Restart your backend server after adding variables

### "Failed to create GitHub issue"
- Check that your GitHub token has the correct permissions
- Verify the repository owner and name are correct
- Ensure the token hasn't expired

### Issues not appearing in GitHub
- Check the backend logs for error messages
- Verify the repository exists and you have write access
- Test the token using GitHub's API directly

## Future Enhancements

Potential improvements for the feedback system:

- [ ] Add screenshot upload capability
- [ ] Include user profile info in issue (optional)
- [ ] Add email notification when issue is resolved
- [ ] Create feedback dashboard for admins
- [ ] Add voting system for popular requests
- [ ] Integrate with project management tools (Jira, Linear, etc.)
