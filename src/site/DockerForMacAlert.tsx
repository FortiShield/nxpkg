import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'

/**
 * A global alert telling all users that due to Docker for Mac, site performance
 * will be degraded.
 */
export const DockerForMacAlert: React.SFC<{ className?: string }> = ({ className = '' }) => (
    <DismissibleAlert
        partialStorageKey="DockerForMac"
        className={`alert-animated-bg alert-warning docker-for-mac-alert d-flex align-items-center ${className}`}
    >
        <span className="docker-for-mac-alert__left">
            It looks like you're using Docker for Mac. Due to known issues related to Docker for Mac's file system
            access, search performance and cloning repositories on Nxpkg will be much slower.
        </span>
        <span className="docker-for-mac-alert__right">
            <a target="_blank" href="https://about.nxpkg.com/docs">
                Run Nxpkg on a different platform or deploy it to a server
            </a>{' '}
            for much faster performance.
        </span>
    </DismissibleAlert>
)
