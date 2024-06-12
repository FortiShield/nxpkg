// This file contains information about the features supported by the Nxpkg server.
//
// TODO(sqs): It should probably be populated directly by the server (and be respected
// by the server) instead of being computed here. That change can be made gradually.

/**
 * Whether the server can list all of its repositories. False for Nxpkg.com,
 * which is a mirror of all public GitHub.com repositories.
 */
export const canListAllRepositories = !window.context.nxpkgDotComMode

/**
 * Whether the application should show the user marketing elements (links, etc.)
 * that are intended for Nxpkg.com.
 */
export const showDotComMarketing = window.context.nxpkgDotComMode

/**
 * Whether the signup form should show terms and privacy policy links.
 */
export const signupTerms = window.context.nxpkgDotComMode
