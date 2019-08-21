import React from 'react'
import { Link } from 'react-router-dom'
import clsx from 'clsx'
import Drawer from '@material-ui/core/Drawer'
import List from '@material-ui/core/List'
import Divider from '@material-ui/core/Divider'
import IconButton from '@material-ui/core/IconButton'
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import DashboardIcon from '@material-ui/icons/Dashboard'
import BarChartIcon from '@material-ui/icons/BarChart'

import { useStyles } from 'components/App/Private/styles'
import { privateHomePath, privateInstancesPath } from 'navigation'

const ListItemLink = ({ icon, primary, to }) => {
  const renderLink = React.forwardRef((itemProps, ref) => (<Link to={to} {...itemProps} innerRef={ref} />))
  return (
    <li>
      <ListItem button component={renderLink}>
        <ListItemIcon>{icon}</ListItemIcon>
        <ListItemText primary={primary} />
      </ListItem>
    </li>
  )
}

const MenuItems = () => (
  <div>
    <ListItemLink to={privateHomePath} primary="Dashboard" icon={<BarChartIcon />} />
    <ListItemLink to={privateInstancesPath} primary="Instances" icon={<DashboardIcon />} />
  </div>
)

const Menu = ({ open, handleDrawerClose }) => {
  const classes = useStyles()
  return (
    <Drawer
      variant="permanent"
      classes={{
        paper: clsx(classes.drawerPaper, !open && classes.drawerPaperClose),
      }}
      open={open}
    >
      <div className={classes.toolbarIcon}>
        <IconButton onClick={handleDrawerClose}>
          <ChevronLeftIcon />
        </IconButton>
      </div>
      <Divider />
      <List>
        <MenuItems />
      </List>
    </Drawer>
  )
}

export default Menu
