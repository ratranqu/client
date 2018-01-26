package systests

import (
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol/keybase1"
	"github.com/keybase/client/go/teams"
	"github.com/stretchr/testify/require"
)

func TestTeamOpenAutoAddMember(t *testing.T) {
	tt := newTeamTester(t)
	defer tt.cleanup()

	own := tt.addUser("own")
	roo := tt.addUser("roo")

	teamName, err := libkb.RandString("tt", 5)
	require.NoError(t, err)
	teamName = strings.ToLower(teamName)

	cli := own.teamsClient
	createRes, err := cli.TeamCreateWithSettings(context.TODO(), keybase1.TeamCreateWithSettingsArg{
		Name: teamName,
		Settings: keybase1.TeamSettings{
			Open:   true,
			JoinAs: keybase1.TeamRole_READER,
		},
	})
	require.NoError(t, err)
	teamID := createRes.TeamID

	t.Logf("Open team name is %q", teamName)

	ret, err := roo.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: teamName})
	require.NoError(t, err)
	require.Equal(t, true, ret.Open)

	own.kickTeamRekeyd()
	own.waitForTeamChangedGregor(teamID, keybase1.Seqno(2))

	teamObj, err := teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
		Name:        teamName,
		ForceRepoll: true,
	})
	require.NoError(t, err)

	role, err := teamObj.MemberRole(context.TODO(), roo.userVersion())
	require.NoError(t, err)
	require.Equal(t, role, keybase1.TeamRole_READER)
}

func TestTeamOpenSettings(t *testing.T) {
	tt := newTeamTester(t)
	defer tt.cleanup()

	own := tt.addUser("own")

	teamName := own.createTeam()
	t.Logf("Open team name is %q", teamName)

	loadTeam := func() *teams.Team {
		ret, err := teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
			Name:        teamName,
			ForceRepoll: true,
		})
		require.NoError(t, err)
		return ret
	}

	teamObj := loadTeam()
	require.Equal(t, teamObj.IsOpen(), false)

	err := teams.ChangeTeamSettings(context.TODO(), own.tc.G, teamName, keybase1.TeamSettings{Open: true, JoinAs: keybase1.TeamRole_READER})
	require.NoError(t, err)

	teamObj = loadTeam()
	require.Equal(t, teamObj.IsOpen(), true)

	err = teams.ChangeTeamSettings(context.TODO(), own.tc.G, teamName, keybase1.TeamSettings{Open: false})
	require.NoError(t, err)

	teamObj = loadTeam()
	require.Equal(t, teamObj.IsOpen(), false)
}

func TestOpenSubteamAdd(t *testing.T) {
	tt := newTeamTester(t)
	defer tt.cleanup()

	own := tt.addUser("own")
	roo := tt.addUser("roo")

	// Creating team, subteam, sending open setting, checking if it's set.

	team := own.createTeam()

	parentName, err := keybase1.TeamNameFromString(team)
	require.NoError(t, err)

	subteam, err := teams.CreateSubteam(context.TODO(), own.tc.G, "zzz", parentName)
	require.NoError(t, err)

	t.Logf("Open team name is %q, subteam is %q", team, subteam)

	subteamObj, err := teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
		ID:          *subteam,
		ForceRepoll: true,
	})
	require.NoError(t, err)

	err = teams.ChangeTeamSettings(context.TODO(), own.tc.G, subteamObj.Name().String(), keybase1.TeamSettings{Open: true, JoinAs: keybase1.TeamRole_READER})
	require.NoError(t, err)

	subteamObj, err = teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
		ID:          *subteam,
		ForceRepoll: true,
	})
	require.NoError(t, err)
	require.Equal(t, subteamObj.IsOpen(), true)

	// User requesting access
	subteamNameStr := subteamObj.Name().String()
	roo.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: subteamNameStr})

	own.kickTeamRekeyd()
	own.waitForTeamChangedGregor(*subteam, keybase1.Seqno(3))

	subteamObj, err = teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
		ID:          *subteam,
		ForceRepoll: true,
	})
	require.NoError(t, err)

	role, err := subteamObj.MemberRole(context.TODO(), roo.userVersion())
	require.NoError(t, err)
	require.Equal(t, role, keybase1.TeamRole_READER)
}

func TestTeamOpenMultipleTars(t *testing.T) {
	tt := newTeamTester(t)
	defer tt.cleanup()

	tar1 := tt.addUser("roo1")
	tar2 := tt.addUser("roo2")
	tar3 := tt.addUser("roo3")
	own := tt.addUser("own")

	teamID, teamName := own.createTeam2()
	t.Logf("Open team name is %q", teamName.String())

	// tar1 and tar2 request access before team is open.
	tar1.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: teamName.String()})
	tar2.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: teamName.String()})

	// Change settings to open
	err := teams.ChangeTeamSettings(context.TODO(), own.tc.G, teamName.String(), keybase1.TeamSettings{Open: true, JoinAs: keybase1.TeamRole_READER})
	require.NoError(t, err)

	// tar3 requests, but rekeyd will grab all requests
	tar3.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: teamName.String()})

	own.kickTeamRekeyd()
	own.waitForTeamChangedGregor(teamID, keybase1.Seqno(3))

	teamObj, err := teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
		Name:        teamName.String(),
		ForceRepoll: true,
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		role, err := teamObj.MemberRole(context.TODO(), tt.users[i].userVersion())
		require.NoError(t, err)
		require.Equal(t, role, keybase1.TeamRole_READER)
	}
}

func TestTeamOpenBans(t *testing.T) {
	tt := newTeamTester(t)
	defer tt.cleanup()

	own := tt.addUser("own")
	bob := tt.addUser("bob")

	team := own.createTeam()
	t.Logf("Open team name is %q", team)

	teamName, err := keybase1.TeamNameFromString(team)
	require.NoError(t, err)

	t.Logf("Trying team edit cli...")
	runner := client.NewCmdTeamSettingsRunner(own.tc.G)
	runner.Team = teamName
	joinAsRole := keybase1.TeamRole_READER
	runner.JoinAsRole = &joinAsRole
	err = runner.Run()
	require.NoError(t, err)

	own.addTeamMember(team, bob.username, keybase1.TeamRole_READER)

	removeRunner := client.NewCmdTeamRemoveMemberRunner(own.tc.G)
	removeRunner.Team = team
	removeRunner.Username = bob.username
	removeRunner.Force = true
	err = removeRunner.Run()
	require.NoError(t, err)

	_, err = bob.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: team})
	require.Error(t, err)
	appErr, ok := err.(libkb.AppStatusError)
	require.True(t, ok)
	require.Equal(t, appErr.Code, libkb.SCTeamBanned)
}

func TestTeamOpenPuklessRequest(t *testing.T) {
	t.Skip() // See CORE-6841

	tt := newTeamTester(t)
	defer tt.cleanup()

	own := tt.addUser("own")
	bob := tt.addPuklessUser("bob")

	team := own.createTeam()
	t.Logf("Open team name is %q", team)

	err := teams.ChangeTeamSettings(context.TODO(), own.tc.G, team, keybase1.TeamSettings{Open: true, JoinAs: keybase1.TeamRole_READER})
	require.NoError(t, err)

	_, err = bob.teamsClient.TeamRequestAccess(context.TODO(), keybase1.TeamRequestAccessArg{Name: team})
	require.NoError(t, err)

	own.kickTeamRekeyd()
	own.pollForTeamSeqnoLink(team, keybase1.Seqno(3))

	teamObj, err := teams.Load(context.TODO(), own.tc.G, keybase1.LoadTeamArg{
		Name:        team,
		ForceRepoll: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, teamObj.NumActiveInvites())

	members, err := teamObj.Members()
	require.NoError(t, err)
	require.Equal(t, 1, len(members.AllUIDs())) // just owner
}

func TestTeamOpenRemoveOldUVAddInvite(t *testing.T) {
	t.Skip() // See CORE-6841
	ctx := newSMUContext(t)
	defer ctx.cleanup()

	ann := ctx.installKeybaseForUser("ann", 10)
	ann.signup()
	bob := ctx.installKeybaseForUser("bob", 10)
	bob.signup()

	team := ann.createTeam([]*smuUser{bob})
	t.Logf("Open team name is %q", team)

	annCtx := ann.getPrimaryGlobalContext()

	ann.openTeam(team, keybase1.TeamRole_READER)

	bob.reset()
	bob.loginAfterResetNoPUK(10)

	bob.requestAccess(team)

	kickTeamRekeyd(annCtx, t)
	ann.pollForTeamSeqnoLink(team, keybase1.Seqno(6))

	teamObj, err := teams.Load(context.TODO(), annCtx, keybase1.LoadTeamArg{
		Name:        team.name,
		ForceRepoll: true,
	})
	require.NoError(t, err)

	loadUserArg := libkb.NewLoadUserArg(annCtx).
		WithNetContext(context.TODO()).
		WithName(bob.username).
		WithPublicKeyOptional().
		WithForcePoll(true)
	upak, _, err := annCtx.GetUPAKLoader().LoadV2(loadUserArg)
	require.NoError(t, err)

	seqno := upak.Current.EldestSeqno

	require.Equal(t, 1, teamObj.NumActiveInvites())
	ret, err := teamObj.HasActiveInvite(teams.NewUserVersion(bob.uid(), seqno).TeamInviteName(), "keybase")
	require.NoError(t, err)
	require.True(t, ret)

	members, err := teamObj.Members()
	require.NoError(t, err)
	// expecting just ann, pre-reset version of bob should have been removed.
	require.Equal(t, 1, len(members.AllUIDs()))
	require.Equal(t, ann.uid(), members.AllUIDs()[0])
}

// Consider user that resets their account and immediately tries to
// re-join their open teams from the website, before provisioning (and
// therefore getting a PUK).
func TestTeamOpenResetAndRejoin(t *testing.T) {
	t.Skip() // See CORE-6841
	ctx := newSMUContext(t)
	defer ctx.cleanup()

	ann := ctx.installKeybaseForUser("ann", 10)
	ann.signup()
	bob := ctx.installKeybaseForUser("bob", 10)
	bob.signup()

	team := ann.createTeam([]*smuUser{bob})
	t.Logf("Open team name is %q", team)

	ann.openTeam(team, keybase1.TeamRole_READER)

	annCtx := ann.getPrimaryGlobalContext()
	bobCtx := bob.getPrimaryGlobalContext()

	// Bob is in the team but he resets and doesn't provision.
	bob.reset()

	loadUserArg := libkb.NewLoadUserArg(annCtx).
		WithNetContext(context.TODO()).
		WithName(bob.username).
		WithPublicKeyOptional().
		WithForcePoll(true)
	upak, _, err := annCtx.GetUPAKLoader().LoadV2(loadUserArg)
	require.NoError(t, err)

	// His EldestSeqno is 0 (in the middle of reset).
	require.EqualValues(t, 0, upak.Current.EldestSeqno)

	// Then bob makes a team access request (hypothetically from the
	// website using "Join team" button).
	err = bobCtx.LoginState().LoginWithPassphrase(bob.username, bob.userInfo.passphrase, false, nil)
	require.NoError(t, err)

	arg := libkb.NewAPIArgWithNetContext(context.TODO(), "team/request_access")
	arg.Args = libkb.NewHTTPArgs()
	arg.SessionType = libkb.APISessionTypeREQUIRED
	arg.Args.Add("team", libkb.S{Val: team.name})
	_, err = bobCtx.API.Post(arg)
	require.NoError(t, err)

	// We are expected to see following new links:
	// - Rotate key (after bob resets)
	// - Change membership (remove reset version of bob)
	// - Invite (add bob%0 as keybase-type invite)
	kickTeamRekeyd(annCtx, t)
	ann.pollForTeamSeqnoLink(team, keybase1.Seqno(6))

	loadTeamArg := keybase1.LoadTeamArg{
		Name:        team.name,
		ForceRepoll: true,
	}
	teamObj, err := teams.Load(context.TODO(), annCtx, loadTeamArg)
	require.NoError(t, err)

	require.Equal(t, 1, teamObj.NumActiveInvites())
	// Expect to see bob%0 keybase invite.
	ret, err := teamObj.HasActiveInvite(teams.NewUserVersion(bob.uid(), 0).TeamInviteName(), "keybase")
	require.NoError(t, err)
	require.True(t, ret)

	members, err := teamObj.Members()
	require.NoError(t, err)
	// expecting just ann, pre-reset version of bob should have been removed.
	require.Equal(t, 1, len(members.AllUIDs()))
	require.Equal(t, ann.uid(), members.AllUIDs()[0])

	// Finally bob gets a PUK - should be automatically added by SBS
	// handler.
	bob.loginAfterReset(10)

	// We are expecting a new ChangeMembership link that adds bob.
	kickTeamRekeyd(annCtx, t)
	ann.pollForTeamSeqnoLink(team, keybase1.Seqno(7))

	teamObj, err = teams.Load(context.TODO(), annCtx, loadTeamArg)
	require.NoError(t, err)

	require.Equal(t, 0, teamObj.NumActiveInvites())

	members, err = teamObj.Members()
	require.NoError(t, err)
	// Expecting bob and ann finally reunited as proper cryptomembers.
	require.Equal(t, 2, len(members.AllUIDs()))
}
