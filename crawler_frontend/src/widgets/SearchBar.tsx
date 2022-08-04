import styles from './SearchBar.module.scss';

export interface SearchBarProps {
    placeholder: string
    value: string
    onChange: (value: string) => void
}

export const SearchBar = (props: SearchBarProps) => {
    return <div className={ styles.search_bar }>
        <input placeholder={ props.placeholder } value={ props.value } onChange={ e => props.onChange( e.target.value )}></input>
    </div>
}