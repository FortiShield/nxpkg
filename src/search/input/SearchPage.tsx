import * as H from 'history'
import * as React from 'react'
import { parseSearchURLQuery } from '..'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { OpenHelpPopoverButton } from '../../global/OpenHelpPopoverButton'
import { eventLogger } from '../../tracking/eventLogger'
import { limitString } from '../../util'
import { queryIndexOfScope, submitSearch } from '../helpers'
import { SavedQueries } from '../saved-queries/SavedQueries'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { SearchFilterChips } from './SearchFilterChips'

interface Props {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onThemeChange: () => void
    showHelpPopover: boolean
    onHelpPopoverToggle: (visible?: boolean) => void
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, State> {
    private static HIDE_REPOGROUP_SAMPLE_STORAGE_KEY = 'SearchPage/hideRepogroupSample'

    constructor(props: Props) {
        super(props)

        const searchOptions = parseSearchURLQuery(props.location.search)
        this.state = {
            userQuery: (searchOptions && searchOptions.query) || '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')
        if (
            window.context.nxpkgDotComMode &&
            !localStorage.getItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY) &&
            !this.state.userQuery
        ) {
            this.setState({ userQuery: 'repogroup:sample' })
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-page">
                <PageTitle title={this.getPageTitle()} />
                <img
                    className="search-page__logo"
                    src={
                        `${window.context.assetsRoot}/img/nxpkg` +
                        (this.props.isLightTheme ? '-light' : '') +
                        '-head-logo.svg'
                    }
                />
                <Form className="search search-page__container" onSubmit={this.onSubmit}>
                    <div className="search-page__input-container">
                        <QueryInput
                            {...this.props}
                            value={this.state.userQuery}
                            onChange={this.onUserQueryChange}
                            autoFocus={'cursor-at-end'}
                            hasGlobalQueryBehavior={true}
                        />
                        <SearchButton />
                        <OpenHelpPopoverButton onHelpPopoverToggle={this.props.onHelpPopoverToggle} />
                    </div>
                    <div className="search-page__input-sub-container">
                        <SearchFilterChips
                            location={this.props.location}
                            history={this.props.history}
                            query={this.state.userQuery}
                        />
                    </div>
                </Form>
                <div className="search search-page__query-container">
                    <SavedQueries {...this.props} hideExampleSearches={true} />
                </div>
            </div>
        )
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })

        if (window.context.nxpkgDotComMode) {
            if (queryIndexOfScope(userQuery, 'repogroup:sample') !== -1) {
                localStorage.removeItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
            } else {
                localStorage.setItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
            }
        }
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(this.props.history, { query: this.state.userQuery }, 'home')
    }

    private getPageTitle(): string | undefined {
        const options = parseSearchURLQuery(this.props.location.search)
        if (options && options.query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
