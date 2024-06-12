import { ConfiguredExtension } from '@nxpkg/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascade,
    ConfigurationSubject,
    Settings,
} from '@nxpkg/extensions-client-common/lib/settings'
import { Environment } from 'nxpkg/module/client/environment'
import { TextDocumentItem } from 'nxpkg/module/client/types/textDocument'

/** React props or state representing the Nxpkg extensions environment. */
export interface ExtensionsEnvironmentProps {
    /** The Nxpkg extensions environment. */
    extensionsEnvironment: Environment<ConfiguredExtension, ConfigurationCascade<ConfigurationSubject, Settings>>
}

/** React props for components that participate in the Nxpkg extensions environment. */
export interface ExtensionsDocumentsProps {
    /**
     * Called when the Nxpkg extensions environment's set of visible text documents changes.
     */
    extensionsOnVisibleTextDocumentsChange: (visibleTextDocuments: TextDocumentItem[] | null) => void
}
