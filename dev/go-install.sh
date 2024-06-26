#!/usr/bin/env bash
#
# build commands, optionally with or without race detector.
# a list of every command we know about, to use by default
all_commands=" gitserver indexer query-runner github-proxy lsp-proxy searcher frontend repo-updater symbols "

# GOMOD_ROOT is the directory from which `go install` commands are run. It should contain a go.mod
# file. The go.mod file may be updated as a side effect of updating the dependencies before the `go
# install`.
GOMOD_ROOT=${GOMOD_ROOT:-$PWD}
echo >&2 "Running \`go install\` from $GOMOD_ROOT"

# handle options
verbose=false
while getopts 'v' o; do
	case $o in
	v)	verbose=true;;
	\?)	echo >&2 "usage: go-install.sh [-v] [commands]"
		exit 1
		;;
	esac
done
shift $(expr $OPTIND - 1)

# check provided commands
ok=true
case $# in
0)	commands=$all_commands;;
*)	commands=" $* "
	for cmd in $commands; do
		case $all_commands in
		*" $cmd "*)	;;
		*)	echo >&2 "unknown command: $cmd"
			ok=false
			;;
		esac
	done
	;;
esac

$ok || exit 1

mkdir -p .bin
export GOBIN=$PWD/.bin
export GO111MODULE=on

if ! go install \
	github.com/mattn/goreman \
	github.com/google/zoekt/cmd/zoekt-archive-index \
	github.com/google/zoekt/cmd/zoekt-nxpkg-indexserver \
	github.com/google/zoekt/cmd/zoekt-webserver; then
	echo >&2 "failed to install prerequisites, aborting."
	exit 1
fi

TAGS='dev'
if [ -n "$DELVE" ]; then
	echo >&2 'Building with optimizations disabled (for debugging). Make sure you have at least go1.10 installed.'
	GCFLAGS='all=-N -l'
	TAGS="$TAGS delve"
fi

# build a list of "cmd,true" and "cmd,false" pairs to indicate whether each command
# wants its own flags. we can't use variable names with the command in them because
# some commands have hyphens.
raced=""
unraced=""
case $GORACED in
"all")	for cmd in $commands; do
		raced="$raced $cmd"
	done
	;;
*)	for cmd in $commands; do
		case " $GORACED " in
		*" $cmd "*)
			raced="$raced $cmd"
			;;
		*)
			unraced="$unraced $cmd"
			;;
		esac
	done
	;;
esac

# Shared logic for the go install part
do_install() {
	race=$1
	shift
	cmdlist="$*"
	cmds=""
	for cmd in $cmdlist; do
		replaced=false
    		for enterpriseCmd in $ENTERPRISE_COMMANDS; do
			if [ "$cmd" == "$enterpriseCmd" ]; then
				cmds="$cmds github.com/nxpkg/enterprise/cmd/$enterpriseCmd"
				replaced=true
			fi
		done
		if [ $replaced == false ]; then
			cmds="$cmds github.com/nxpkg/nxpkg/cmd/$cmd"
		fi
	done
	if ( cd $GOMOD_ROOT && go install -v -gcflags="$GCFLAGS" -tags "$TAGS" -race=$race $cmds ); then
		if $verbose; then
			# echo each command on its own line
			echo "$cmdlist" | tr ' ' '\012'
		fi
	else
		failed="$failed $cmdlist"
	fi
}

if [ -n "$raced" ]; then
	echo >&2 "Go race detector enabled for: $GORACED."
	do_install true $raced
else
	echo >&2 "Go race detector disabled. You can enable it for specific commands by setting GORACED (e.g. GORACED=frontend,searcher or GORACED=all for all commands)"
fi

if [ -n "$unraced" ]; then
	do_install false $unraced
fi

if [ -n "$failed" ]; then
	echo >&2 "failed to build:$failed"
	exit 1
fi
