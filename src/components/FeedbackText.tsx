import * as React from 'react'

interface Props {
    headerText?: React.ReactNode
    footerText?: React.ReactNode
    className?: string
}

export const FeedbackText: React.StatelessComponent<Props> = (props: Props) => (
    <p className={`feedback-text ${props.className || ''}`}>
        {props.headerText || 'Questions/feedback?'} Contact us at{' '}
        <a href="https://twitter.com/srcgraph" target="_blank">
            @srcgraph
        </a>{' '}
        or{' '}
        <a href="mailto:support@nxpkg.com" target="_blank">
            support@nxpkg.com
        </a>, or file issues on our{' '}
        <a href="https://github.com/nxpkg/issues/issues" target="_blank">
            public issue tracker
        </a>. {props.footerText}
    </p>
)
