package github

// GetOwner returns the owner for Issue.
func (i *Issue) GetOwner() string { return i.Owner }

// GetRepo returns the repo for Issue.
func (i *Issue) GetRepo() string { return i.Repo }

// GetNumber returns the number for Issue.
func (i *Issue) GetNumber() int { return i.Number }

// GetOwner returns the owner for PullRequest.
func (p *PullRequest) GetOwner() string { return p.Owner }

// GetRepo returns the repo for PullRequest.
func (p *PullRequest) GetRepo() string { return p.Repo }

// GetNumber returns the number for PullRequest.
func (p *PullRequest) GetNumber() int { return p.Number }

// GetOwner returns the owner for Discussion.
func (d *Discussion) GetOwner() string { return d.Owner }

// GetRepo returns the repo for Discussion.
func (d *Discussion) GetRepo() string { return d.Repo }

// GetNumber returns the number for Discussion.
func (d *Discussion) GetNumber() int { return d.Number }
